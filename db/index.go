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
	case client.FieldKind_NILLABLE_STRING, client.FieldKind_FOREIGN_OBJECT:
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

func (i *collectionBaseIndex) getDocFieldValues(doc *client.Document) ([]*client.FieldValue, error) {
	result := make([]*client.FieldValue, 0, len(i.fieldsDescs))
	for iter := range i.fieldsDescs {
		fieldVal, err := doc.TryGetValue(i.fieldsDescs[iter].Name)
		if err != nil {
			return nil, err
		}
		if fieldVal == nil || fieldVal.Value() == nil {
			result = append(result, client.NewFieldValue(client.NONE_CRDT, nil, i.fieldsDescs[iter].Kind))
			continue
		}
		result = append(result, fieldVal)
	}
	return result, nil
}

func (iter *collectionBaseIndex) getDocumentsIndexKey(
	doc *client.Document,
) (core.IndexDataStoreKey, error) {
	fieldValues, err := iter.getDocFieldValues(doc)
	if err != nil {
		return core.IndexDataStoreKey{}, err
	}

	indexDataStoreKey := core.IndexDataStoreKey{}
	indexDataStoreKey.CollectionID = iter.collection.ID()
	indexDataStoreKey.IndexID = iter.desc.ID
	indexDataStoreKey.Fields = make([]core.IndexedField, len(iter.fieldsDescs))
	for i := range iter.fieldsDescs {
		indexDataStoreKey.Fields[i].ID = iter.fieldsDescs[i].ID
		indexDataStoreKey.Fields[i].Value = fieldValues[i]
		indexDataStoreKey.Fields[i].Descending = iter.desc.Fields[i].Descending
	}
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

	key.Fields = append(key.Fields, core.IndexedField{
		ID:    client.FieldID(core.DocIDFieldIndex),
		Value: client.NewFieldValue(client.NONE_CRDT, doc.ID().String(), client.FieldKind_DocID)},
	)
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
	keyBytes, err := core.EncodeIndexDataStoreKey(nil, &key)
	if err != nil {
		return err
	}
	err = txn.Datastore().Put(ctx, ds.NewKey(string(keyBytes)), []byte{})
	if err != nil {
		return NewErrFailedToStoreIndexedField(string(keyBytes), err)
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

func (i *collectionUniqueIndex) save(
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

func (i *collectionUniqueIndex) Save(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	key, val, err := i.prepareIndexRecordToStore(ctx, txn, doc)
	if err != nil {
		return err
	}
	return i.save(ctx, txn, &key, val)
}

func (i *collectionUniqueIndex) newUniqueIndexError(
	doc *client.Document,
) error {
	kvs := make([]errors.KV, 0, len(i.fieldsDescs))
	for iter := range i.fieldsDescs {
		fieldVal, err := doc.TryGetValue(i.fieldsDescs[iter].Name)
		var val any
		if err != nil {
			return err
		}
		// If fieldVal is nil, we leave `val` as is (e.g. nil)
		if fieldVal != nil {
			val = fieldVal.Value()
		}
		kvs = append(kvs, errors.NewKV(i.fieldsDescs[iter].Name, val))
	}

	return NewErrCanNotIndexNonUniqueFields(doc.ID().String(), kvs...)
}

func (i *collectionUniqueIndex) getDocumentsIndexRecord(
	doc *client.Document,
) (core.IndexDataStoreKey, []byte, error) {
	key, err := i.getDocumentsIndexKey(doc)
	if err != nil {
		return core.IndexDataStoreKey{}, nil, err
	}
	if hasIndexKeyNilField(&key) {
		key.Fields = append(key.Fields, core.IndexedField{
			ID:    client.FieldID(core.DocIDFieldIndex),
			Value: client.NewFieldValue(client.NONE_CRDT, doc.ID().String(), client.FieldKind_DocID)},
		)
		return key, []byte{}, nil
	} else {
		return key, []byte(doc.ID().String()), nil
	}
}

func (i *collectionUniqueIndex) prepareIndexRecordToStore(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) (core.IndexDataStoreKey, []byte, error) {
	key, val, err := i.getDocumentsIndexRecord(doc)
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
			return core.IndexDataStoreKey{}, nil, i.newUniqueIndexError(doc)
		}
	}
	return key, val, nil
}

func (i *collectionUniqueIndex) Update(
	ctx context.Context,
	txn datastore.Txn,
	oldDoc *client.Document,
	newDoc *client.Document,
) error {
	newKey, newVal, err := i.prepareIndexRecordToStore(ctx, txn, newDoc)
	if err != nil {
		return err
	}
	err = i.deleteDocIndex(ctx, txn, oldDoc)
	if err != nil {
		return err
	}
	return i.save(ctx, txn, &newKey, newVal)
}

func (i *collectionUniqueIndex) deleteDocIndex(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	key, _, err := i.getDocumentsIndexRecord(doc)
	if err != nil {
		return err
	}
	return i.deleteIndexKey(ctx, txn, key)
}
