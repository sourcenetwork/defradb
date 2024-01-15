// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/request/graphql/schema/types"
)

// CollectionIndex is an interface for collection indexes
// It abstracts away common index functionality to be implemented
// by different index types: non-unique, unique, and composite
type CollectionIndex interface {
	// Save indexes a document by storing it
	Save(context.Context, datastore.Txn, *client.Document) error
	// Update updates an existing document in the index
	Update(context.Context, datastore.Txn, *client.Document, *client.Document) error
	// RemoveAll removes all documents from the index
	RemoveAll(context.Context, datastore.Txn) error
	// Name returns the name of the index
	Name() string
	// Description returns the description of the index
	Description() client.IndexDescription
}

func canConvertIndexFieldValue[T any](val any) bool {
	_, ok := val.(T)
	return ok
}

func getValidateIndexFieldFunc(kind client.FieldKind) func(any) bool {
	switch kind {
	case client.FieldKind_STRING, client.FieldKind_FOREIGN_OBJECT:
		return canConvertIndexFieldValue[string]
	case client.FieldKind_INT:
		return canConvertIndexFieldValue[int64]
	case client.FieldKind_FLOAT:
		return canConvertIndexFieldValue[float64]
	case client.FieldKind_BOOL:
		return canConvertIndexFieldValue[bool]
	case client.FieldKind_BLOB:
		return func(val any) bool {
			blobStrVal, ok := val.(string)
			if !ok {
				return false
			}
			return types.BlobPattern.MatchString(blobStrVal)
		}
	case client.FieldKind_DATETIME:
		return func(val any) bool {
			timeStrVal, ok := val.(string)
			if !ok {
				return false
			}
			_, err := time.Parse(time.RFC3339, timeStrVal)
			return err == nil
		}
	default:
		return nil
	}
}

func getFieldValidateFunc(kind client.FieldKind) (func(any) bool, error) {
	validateFunc := getValidateIndexFieldFunc(kind)
	if validateFunc == nil {
		return nil, NewErrUnsupportedIndexFieldType(kind)
	}
	return validateFunc, nil
}

// NewCollectionIndex creates a new collection index
func NewCollectionIndex(
	collection client.Collection,
	desc client.IndexDescription,
) (CollectionIndex, error) {
	if len(desc.Fields) == 0 {
		return nil, NewErrIndexDescHasNoFields(desc)
	}
	base := collectionBaseIndex{collection: collection, desc: desc}
	base.validateFieldFuncs = make([]func(any) bool, 0, len(desc.Fields))
	base.fieldsDescs = make([]client.FieldDescription, 0, len(desc.Fields))
	for _, fieldDesc := range desc.Fields {
		field, foundField := collection.Schema().GetField(fieldDesc.Name)
		if !foundField {
			return nil, NewErrIndexDescHasNonExistingField(desc, desc.Fields[0].Name)
		}
		base.fieldsDescs = append(base.fieldsDescs, field)
		validateFunc, err := getFieldValidateFunc(field.Kind)
		if err != nil {
			return nil, err
		}
		base.validateFieldFuncs = append(base.validateFieldFuncs, validateFunc)
	}
	if desc.Unique {
		return &collectionUniqueIndex{collectionBaseIndex: base}, nil
	} else {
		return &collectionSimpleIndex{collectionBaseIndex: base}, nil
	}
}

type collectionBaseIndex struct {
	collection         client.Collection
	desc               client.IndexDescription
	validateFieldFuncs []func(any) bool
	fieldsDescs        []client.FieldDescription
}

func (i *collectionBaseIndex) getDocFieldValue(doc *client.Document) ([][]byte, error) {
	result := make([][]byte, 0, len(i.fieldsDescs))
	for iter := range i.fieldsDescs {
		fieldVal, err := doc.GetValue(i.fieldsDescs[iter].Name)
		if err != nil {
			if errors.Is(err, client.ErrFieldNotExist) {
				valBytes, err := client.NewFieldValue(client.LWW_REGISTER, nil).Bytes()
				if err != nil {
					return nil, err
				}
				result = append(result, valBytes)
				continue
			}
			return nil, err
		}
		if !i.validateFieldFuncs[iter](fieldVal.Value()) {
			return nil, NewErrInvalidFieldValue(i.fieldsDescs[iter].Kind, fieldVal)
		}
		valBytes, err := fieldVal.Bytes()
		if err != nil {
			return nil, err
		}
		result = append(result, valBytes)
	}
	return result, nil
}

func (i *collectionBaseIndex) getDocumentsIndexKey(
	doc *client.Document,
) (core.IndexDataStoreKey, error) {
	fieldValues, err := i.getDocFieldValue(doc)
	if err != nil {
		return core.IndexDataStoreKey{}, err
	}

	indexDataStoreKey := core.IndexDataStoreKey{}
	indexDataStoreKey.CollectionID = i.collection.ID()
	indexDataStoreKey.IndexID = i.desc.ID
	indexDataStoreKey.FieldValues = fieldValues
	return indexDataStoreKey, nil
}

func (i *collectionBaseIndex) deleteIndexKey(
	ctx context.Context,
	txn datastore.Txn,
	key core.IndexDataStoreKey,
) error {
	exists, err := txn.Datastore().Has(ctx, key.ToDS())
	if err != nil {
		return err
	}
	if !exists {
		return NewErrCorruptedIndex(i.desc.Name)
	}
	return txn.Datastore().Delete(ctx, key.ToDS())
}

// RemoveAll remove all artifacts of the index from the storage, i.e. all index
// field values for all documents.
func (i *collectionBaseIndex) RemoveAll(ctx context.Context, txn datastore.Txn) error {
	prefixKey := core.IndexDataStoreKey{}
	prefixKey.CollectionID = i.collection.ID()
	prefixKey.IndexID = i.desc.ID

	keys, err := datastore.FetchKeysForPrefix(ctx, prefixKey.ToString(), txn.Datastore())
	if err != nil {
		return err
	}

	for _, key := range keys {
		err := txn.Datastore().Delete(ctx, key)
		if err != nil {
			return NewCanNotDeleteIndexedField(err)
		}
	}

	return nil
}

// Name returns the name of the index
func (i *collectionBaseIndex) Name() string {
	return i.desc.Name
}

// Description returns the description of the index
func (i *collectionBaseIndex) Description() client.IndexDescription {
	return i.desc
}

// collectionSimpleIndex is an non-unique index that indexes documents by a single field.
// Single-field indexes store values only in ascending order.
type collectionSimpleIndex struct {
	collectionBaseIndex
}

var _ CollectionIndex = (*collectionSimpleIndex)(nil)

func (i *collectionSimpleIndex) getDocumentsIndexKey(
	doc *client.Document,
) (core.IndexDataStoreKey, error) {
	key, err := i.collectionBaseIndex.getDocumentsIndexKey(doc)
	if err != nil {
		return core.IndexDataStoreKey{}, err
	}

	key.FieldValues = append(key.FieldValues, []byte(doc.ID().String()))
	return key, nil
}

// Save indexes a document by storing the indexed field value.
func (i *collectionSimpleIndex) Save(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	key, err := i.getDocumentsIndexKey(doc)
	if err != nil {
		return err
	}
	err = txn.Datastore().Put(ctx, key.ToDS(), []byte{})
	if err != nil {
		return NewErrFailedToStoreIndexedField(key.ToDS().String(), err)
	}
	return nil
}

func (i *collectionSimpleIndex) Update(
	ctx context.Context,
	txn datastore.Txn,
	oldDoc *client.Document,
	newDoc *client.Document,
) error {
	err := i.deleteDocIndex(ctx, txn, oldDoc)
	if err != nil {
		return err
	}
	return i.Save(ctx, txn, newDoc)
}

func (i *collectionSimpleIndex) deleteDocIndex(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	key, err := i.getDocumentsIndexKey(doc)
	if err != nil {
		return err
	}
	return i.deleteIndexKey(ctx, txn, key)
}

type collectionUniqueIndex struct {
	collectionBaseIndex
}

var _ CollectionIndex = (*collectionUniqueIndex)(nil)

func (i *collectionUniqueIndex) Save(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	key, err := i.getDocumentsIndexKey(doc)
	if err != nil {
		return err
	}
	exists, err := txn.Datastore().Has(ctx, key.ToDS())
	if err != nil {
		return err
	}
	if exists {
		return i.newUniqueIndexError(doc)
	}
	err = txn.Datastore().Put(ctx, key.ToDS(), []byte(doc.ID().String()))
	if err != nil {
		return NewErrFailedToStoreIndexedField(key.ToDS().String(), err)
	}
	return nil
}

func (i *collectionUniqueIndex) newUniqueIndexError(
	doc *client.Document,
) error {
	fieldVal, err := doc.GetValue(i.fieldsDescs[0].Name)
	var val any
	if err != nil {
		// If the error is ErrFieldNotExist, we leave `val` as is (e.g. nil)
		// otherwise we return the error
		if !errors.Is(err, client.ErrFieldNotExist) {
			return err
		}
	} else {
		val = fieldVal.Value()
	}

	return NewErrCanNotIndexNonUniqueField(doc.ID().String(), i.fieldsDescs[0].Name, val)
}

func (i *collectionUniqueIndex) Update(
	ctx context.Context,
	txn datastore.Txn,
	oldDoc *client.Document,
	newDoc *client.Document,
) error {
	newKey, err := i.getDocumentsIndexKey(newDoc)
	if err != nil {
		return err
	}
	exists, err := txn.Datastore().Has(ctx, newKey.ToDS())
	if err != nil {
		return err
	}
	if exists {
		return i.newUniqueIndexError(newDoc)
	}
	err = i.deleteDocIndex(ctx, txn, oldDoc)
	if err != nil {
		return err
	}
	return i.Save(ctx, txn, newDoc)
}

func (i *collectionUniqueIndex) deleteDocIndex(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	key, err := i.getDocumentsIndexKey(doc)
	if err != nil {
		return err
	}
	return i.deleteIndexKey(ctx, txn, key)
}
