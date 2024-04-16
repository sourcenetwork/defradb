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
	client.CollectionIndex
	// RemoveAll removes all documents from the index
	RemoveAll(context.Context, datastore.Txn) error
}

func canConvertIndexFieldValue[T any](val any) bool {
	_, ok := val.(T)
	return ok
}

func getValidateIndexFieldFunc(kind client.FieldKind) func(any) bool {
	if kind.IsObject() && !kind.IsArray() {
		return canConvertIndexFieldValue[string]
	}

	switch kind {
	case client.FieldKind_NILLABLE_STRING:
		return canConvertIndexFieldValue[string]
	case client.FieldKind_NILLABLE_INT:
		return canConvertIndexFieldValue[int64]
	case client.FieldKind_NILLABLE_FLOAT:
		return canConvertIndexFieldValue[float64]
	case client.FieldKind_NILLABLE_BOOL:
		return canConvertIndexFieldValue[bool]
	case client.FieldKind_NILLABLE_BLOB:
		return func(val any) bool {
			blobStrVal, ok := val.(string)
			if !ok {
				return false
			}
			return types.BlobPattern.MatchString(blobStrVal)
		}
	case client.FieldKind_NILLABLE_DATETIME:
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
	base.validateFieldFuncs = make([]func(any) bool, len(desc.Fields))
	base.fieldsDescs = make([]client.SchemaFieldDescription, len(desc.Fields))
	for i := range desc.Fields {
		field, foundField := collection.Schema().GetFieldByName(desc.Fields[i].Name)
		if !foundField {
			return nil, client.NewErrFieldNotExist(desc.Fields[i].Name)
		}
		base.fieldsDescs[i] = field
		validateFunc, err := getFieldValidateFunc(field.Kind)
		if err != nil {
			return nil, err
		}
		base.validateFieldFuncs[i] = validateFunc
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
	fieldsDescs        []client.SchemaFieldDescription
}

func (index *collectionBaseIndex) getDocFieldValues(doc *client.Document) ([]client.NormalValue, error) {
	result := make([]client.NormalValue, 0, len(index.fieldsDescs))
	for iter := range index.fieldsDescs {
		fieldVal, err := doc.TryGetValue(index.fieldsDescs[iter].Name)
		if err != nil {
			return nil, err
		}
		if fieldVal == nil || fieldVal.Value() == nil {
			normalNil, err := client.NewNormalNil(index.fieldsDescs[iter].Kind)
			if err != nil {
				return nil, err
			}
			result = append(result, normalNil)
			continue
		}
		result = append(result, fieldVal.NormalValue())
	}
	return result, nil
}

func (index *collectionBaseIndex) getDocumentsIndexKey(
	doc *client.Document,
) (core.IndexDataStoreKey, error) {
	fieldValues, err := index.getDocFieldValues(doc)
	if err != nil {
		return core.IndexDataStoreKey{}, err
	}

	fields := make([]core.IndexedField, len(index.fieldsDescs))
	for i := range index.fieldsDescs {
		fields[i].Value = fieldValues[i]
		fields[i].Descending = index.desc.Fields[i].Descending
	}
	return core.NewIndexDataStoreKey(index.collection.ID(), index.desc.ID, fields), nil
}

func (index *collectionBaseIndex) deleteIndexKey(
	ctx context.Context,
	txn datastore.Txn,
	key core.IndexDataStoreKey,
) error {
	exists, err := txn.Datastore().Has(ctx, key.ToDS())
	if err != nil {
		return err
	}
	if !exists {
		return NewErrCorruptedIndex(index.desc.Name)
	}
	return txn.Datastore().Delete(ctx, key.ToDS())
}

// RemoveAll remove all artifacts of the index from the storage, i.e. all index
// field values for all documents.
func (index *collectionBaseIndex) RemoveAll(ctx context.Context, txn datastore.Txn) error {
	prefixKey := core.IndexDataStoreKey{}
	prefixKey.CollectionID = index.collection.ID()
	prefixKey.IndexID = index.desc.ID

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
func (index *collectionBaseIndex) Name() string {
	return index.desc.Name
}

// Description returns the description of the index
func (index *collectionBaseIndex) Description() client.IndexDescription {
	return index.desc
}

// collectionSimpleIndex is an non-unique index that indexes documents by a single field.
// Single-field indexes store values only in ascending order.
type collectionSimpleIndex struct {
	collectionBaseIndex
}

var _ CollectionIndex = (*collectionSimpleIndex)(nil)

func (index *collectionSimpleIndex) getDocumentsIndexKey(
	doc *client.Document,
) (core.IndexDataStoreKey, error) {
	key, err := index.collectionBaseIndex.getDocumentsIndexKey(doc)
	if err != nil {
		return core.IndexDataStoreKey{}, err
	}

	key.Fields = append(key.Fields, core.IndexedField{Value: client.NewNormalString(doc.ID().String())})
	return key, nil
}

// Save indexes a document by storing the indexed field value.
func (index *collectionSimpleIndex) Save(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	key, err := index.getDocumentsIndexKey(doc)
	if err != nil {
		return err
	}
	err = txn.Datastore().Put(ctx, key.ToDS(), []byte{})
	if err != nil {
		return NewErrFailedToStoreIndexedField(key.ToString(), err)
	}
	return nil
}

func (index *collectionSimpleIndex) Update(
	ctx context.Context,
	txn datastore.Txn,
	oldDoc *client.Document,
	newDoc *client.Document,
) error {
	err := index.deleteDocIndex(ctx, txn, oldDoc)
	if err != nil {
		return err
	}
	return index.Save(ctx, txn, newDoc)
}

func (index *collectionSimpleIndex) Delete(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	return index.deleteDocIndex(ctx, txn, doc)
}

func (index *collectionSimpleIndex) deleteDocIndex(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	key, err := index.getDocumentsIndexKey(doc)
	if err != nil {
		return err
	}
	return index.deleteIndexKey(ctx, txn, key)
}

// hasIndexKeyNilField returns true if the index key has a field with nil value
func hasIndexKeyNilField(key *core.IndexDataStoreKey) bool {
	for i := range key.Fields {
		if key.Fields[i].Value.IsNil() {
			return true
		}
	}
	return false
}

type collectionUniqueIndex struct {
	collectionBaseIndex
}

var _ CollectionIndex = (*collectionUniqueIndex)(nil)

func (index *collectionUniqueIndex) save(
	ctx context.Context,
	txn datastore.Txn,
	key *core.IndexDataStoreKey,
	val []byte,
) error {
	err := txn.Datastore().Put(ctx, key.ToDS(), val)
	if err != nil {
		return NewErrFailedToStoreIndexedField(key.ToDS().String(), err)
	}
	return nil
}

func (index *collectionUniqueIndex) Save(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	key, val, err := index.prepareIndexRecordToStore(ctx, txn, doc)
	if err != nil {
		return err
	}
	return index.save(ctx, txn, &key, val)
}

func (index *collectionUniqueIndex) newUniqueIndexError(
	doc *client.Document,
) error {
	kvs := make([]errors.KV, 0, len(index.fieldsDescs))
	for iter := range index.fieldsDescs {
		fieldVal, err := doc.TryGetValue(index.fieldsDescs[iter].Name)
		var val any
		if err != nil {
			return err
		}
		// If fieldVal is nil, we leave `val` as is (e.g. nil)
		if fieldVal != nil {
			val = fieldVal.Value()
		}
		kvs = append(kvs, errors.NewKV(index.fieldsDescs[iter].Name, val))
	}

	return NewErrCanNotIndexNonUniqueFields(doc.ID().String(), kvs...)
}

func (index *collectionUniqueIndex) getDocumentsIndexRecord(
	doc *client.Document,
) (core.IndexDataStoreKey, []byte, error) {
	key, err := index.getDocumentsIndexKey(doc)
	if err != nil {
		return core.IndexDataStoreKey{}, nil, err
	}
	if hasIndexKeyNilField(&key) {
		key.Fields = append(key.Fields, core.IndexedField{Value: client.NewNormalString(doc.ID().String())})
		return key, []byte{}, nil
	} else {
		return key, []byte(doc.ID().String()), nil
	}
}

func (index *collectionUniqueIndex) prepareIndexRecordToStore(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) (core.IndexDataStoreKey, []byte, error) {
	key, val, err := index.getDocumentsIndexRecord(doc)
	if err != nil {
		return core.IndexDataStoreKey{}, nil, err
	}
	if len(val) != 0 {
		var exists bool
		exists, err = txn.Datastore().Has(ctx, key.ToDS())
		if err != nil {
			return core.IndexDataStoreKey{}, nil, err
		}
		if exists {
			return core.IndexDataStoreKey{}, nil, index.newUniqueIndexError(doc)
		}
	}
	return key, val, nil
}

func (index *collectionUniqueIndex) Delete(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	return index.deleteDocIndex(ctx, txn, doc)
}

func (index *collectionUniqueIndex) Update(
	ctx context.Context,
	txn datastore.Txn,
	oldDoc *client.Document,
	newDoc *client.Document,
) error {
	// We only need to update the index if one of the indexed fields
	// on the document has been changed.
	if !isUpdatingIndexedFields(index, oldDoc, newDoc) {
		return nil
	}
	newKey, newVal, err := index.prepareIndexRecordToStore(ctx, txn, newDoc)
	if err != nil {
		return err
	}
	err = index.deleteDocIndex(ctx, txn, oldDoc)
	if err != nil {
		return err
	}
	return index.save(ctx, txn, &newKey, newVal)
}

func (index *collectionUniqueIndex) deleteDocIndex(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	key, _, err := index.getDocumentsIndexRecord(doc)
	if err != nil {
		return err
	}
	return index.deleteIndexKey(ctx, txn, key)
}

func isUpdatingIndexedFields(index CollectionIndex, oldDoc, newDoc *client.Document) bool {
	for _, indexedFields := range index.Description().Fields {
		oldVal, getOldValErr := oldDoc.GetValue(indexedFields.Name)
		newVal, getNewValErr := newDoc.GetValue(indexedFields.Name)

		// GetValue will return an error when the field doesn't exist.
		// This will happen for oldDoc only if the field hasn't been set
		// when first creating the document. For newDoc, this will happen
		// only if the field hasn't been set when first creating the document
		// AND the field hasn't been set on the update.
		switch {
		case getOldValErr != nil && getNewValErr != nil:
			continue
		case getOldValErr != nil && getNewValErr == nil:
			return true
		case oldVal.Value() != newVal.Value():
			return true
		}
	}
	return false
}
