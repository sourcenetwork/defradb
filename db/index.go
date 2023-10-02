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

	ds "github.com/ipfs/go-datastore"

	"github.com/ipfs/go-datastore/query"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
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
	case client.FieldKind_STRING:
		return canConvertIndexFieldValue[string]
	case client.FieldKind_INT:
		return canConvertIndexFieldValue[int64]
	case client.FieldKind_FLOAT:
		return canConvertIndexFieldValue[float64]
	case client.FieldKind_BOOL:
		return canConvertIndexFieldValue[bool]
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
	index := &collectionSimpleIndex{collection: collection, desc: desc}
	field, foundField := collection.Schema().GetField(desc.Fields[0].Name)
	if !foundField {
		return nil, NewErrIndexDescHasNonExistingField(desc, desc.Fields[0].Name)
	}
	var e error
	index.fieldDesc = field
	index.validateFieldFunc, e = getFieldValidateFunc(field.Kind)
	return index, e
}

// collectionSimpleIndex is an non-unique index that indexes documents by a single field.
// Single-field indexes store values only in ascending order.
type collectionSimpleIndex struct {
	collection        client.Collection
	desc              client.IndexDescription
	validateFieldFunc func(any) bool
	fieldDesc         client.FieldDescription
}

var _ CollectionIndex = (*collectionSimpleIndex)(nil)

func (i *collectionSimpleIndex) getDocumentsIndexKey(
	doc *client.Document,
) (core.IndexDataStoreKey, error) {
	fieldValue, err := i.getDocFieldValue(doc)
	if err != nil {
		return core.IndexDataStoreKey{}, err
	}

	indexDataStoreKey := core.IndexDataStoreKey{}
	indexDataStoreKey.CollectionID = i.collection.ID()
	indexDataStoreKey.IndexID = i.desc.ID
	indexDataStoreKey.FieldValues = [][]byte{fieldValue, []byte(doc.Key().String())}
	return indexDataStoreKey, nil
}

func (i *collectionSimpleIndex) getDocFieldValue(doc *client.Document) ([]byte, error) {
	// collectionSimpleIndex only supports single field indexes, that's why we
	// can safely access the first field
	indexedFieldName := i.desc.Fields[0].Name
	fieldVal, err := doc.GetValue(indexedFieldName)
	if err != nil {
		if errors.Is(err, client.ErrFieldNotExist) {
			return client.NewCBORValue(client.LWW_REGISTER, nil).Bytes()
		} else {
			return nil, err
		}
	}
	writeableVal, ok := fieldVal.(client.WriteableValue)
	if !ok || !i.validateFieldFunc(fieldVal.Value()) {
		return nil, NewErrInvalidFieldValue(i.fieldDesc.Kind, writeableVal)
	}
	return writeableVal.Bytes()
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

// Update updates indexed field values of an existing document.
// It removes the old document from the index and adds the new one.
func (i *collectionSimpleIndex) Update(
	ctx context.Context,
	txn datastore.Txn,
	oldDoc *client.Document,
	newDoc *client.Document,
) error {
	key, err := i.getDocumentsIndexKey(oldDoc)
	if err != nil {
		return err
	}
	err = txn.Datastore().Delete(ctx, key.ToDS())
	if err != nil {
		return err
	}
	return i.Save(ctx, txn, newDoc)
}

func fetchKeysForPrefix(
	ctx context.Context,
	prefix string,
	storage ds.Read,
) ([]ds.Key, error) {
	q, err := storage.Query(ctx, query.Query{Prefix: prefix})
	if err != nil {
		return nil, err
	}

	keys := make([]ds.Key, 0)
	for res := range q.Next() {
		if res.Error != nil {
			_ = q.Close()
			return nil, res.Error
		}
		keys = append(keys, ds.NewKey(res.Key))
	}
	if err = q.Close(); err != nil {
		return nil, err
	}

	return keys, nil
}

// RemoveAll remove all artifacts of the index from the storage, i.e. all index
// field values for all documents.
func (i *collectionSimpleIndex) RemoveAll(ctx context.Context, txn datastore.Txn) error {
	prefixKey := core.IndexDataStoreKey{}
	prefixKey.CollectionID = i.collection.ID()
	prefixKey.IndexID = i.desc.ID

	keys, err := fetchKeysForPrefix(ctx, prefixKey.ToString(), txn.Datastore())
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
func (i *collectionSimpleIndex) Name() string {
	return i.desc.Name
}

// Description returns the description of the index
func (i *collectionSimpleIndex) Description() client.IndexDescription {
	return i.desc
}
