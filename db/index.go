// Copyright 2022 Democratized Data Foundation
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
	"strconv"
	"time"

	ds "github.com/ipfs/go-datastore"

	"github.com/ipfs/go-datastore/query"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
)

const (
	indexFieldValuePrefix = "v"
	indexFieldNilValue    = "n"
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

func getFieldValConverter(kind client.FieldKind) (func(any) ([]byte, error), error) {
	switch kind {
	case client.FieldKind_STRING:
		return func(val any) ([]byte, error) {
			return []byte(val.(string)), nil
		}, nil
	case client.FieldKind_INT:
		return func(val any) ([]byte, error) {
			intVal, ok := val.(int64)
			if !ok {
				return nil, NewErrInvalidFieldValue(kind, val)
			}
			return []byte(strconv.FormatInt(intVal, 10)), nil
		}, nil
	case client.FieldKind_FLOAT:
		return func(val any) ([]byte, error) {
			floatVal, ok := val.(float64)
			if !ok {
				return nil, NewErrInvalidFieldValue(kind, val)
			}
			return []byte(strconv.FormatFloat(floatVal, 'f', -1, 64)), nil
		}, nil
	case client.FieldKind_BOOL:
		return func(val any) ([]byte, error) {
			boolVal, ok := val.(bool)
			if !ok {
				return nil, NewErrInvalidFieldValue(kind, val)
			}
			var intVal int64 = 0
			if boolVal {
				intVal = 1
			}
			return []byte(strconv.FormatInt(intVal, 10)), nil
		}, nil
	case client.FieldKind_DATETIME:
		return func(val any) ([]byte, error) {
			timeStrVal := val.(string)
			_, err := time.Parse(time.RFC3339, timeStrVal)
			if err != nil {
				return nil, NewErrInvalidFieldValue(kind, val)
			}
			return []byte(timeStrVal), nil
		}, nil
	default:
		return nil, NewErrUnsupportedIndexFieldType(kind)
	}
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
	schema := collection.Description().Schema
	fieldID := client.FieldID(schema.GetFieldKey(desc.Fields[0].Name))
	field, foundField := collection.Description().GetFieldByID(fieldID)
	if fieldID == client.FieldID(0) || !foundField {
		return nil, NewErrIndexDescHasNonExistingField(desc, desc.Fields[0].Name)
	}
	var e error
	index.convertFunc, e = getFieldValConverter(field.Kind)
	return index, e
}

// collectionSimpleIndex is an non-unique index that indexes documents by a single field.
// Single-field indexes store values only in ascending order.
type collectionSimpleIndex struct {
	collection  client.Collection
	desc        client.IndexDescription
	convertFunc func(any) ([]byte, error)
}

var _ CollectionIndex = (*collectionSimpleIndex)(nil)

func (i *collectionSimpleIndex) getDocKey(doc *client.Document) (core.IndexDataStoreKey, error) {
	// collectionSimpleIndex only supports single field indexes, that's why we
	// can safely assume access the first field
	indexedFieldName := i.desc.Fields[0].Name
	fieldVal, err := doc.Get(indexedFieldName)
	isNil := false
	if err != nil {
		isNil = errors.Is(err, client.ErrFieldNotExist)
		if !isNil {
			return core.IndexDataStoreKey{}, nil
		}
	}

	var storeValue []byte
	if isNil {
		storeValue = []byte(indexFieldNilValue)
	} else {
		data, err := i.convertFunc(fieldVal)
		if err != nil {
			return core.IndexDataStoreKey{}, err
		}
		storeValue = []byte(string(indexFieldValuePrefix) + string(data))
	}
	indexDataStoreKey := core.IndexDataStoreKey{}
	indexDataStoreKey.CollectionID = i.collection.ID()
	indexDataStoreKey.IndexID = i.desc.ID
	indexDataStoreKey.FieldValues = [][]byte{storeValue,
		[]byte(string(indexFieldValuePrefix) + doc.Key().String())}
	return indexDataStoreKey, nil
}

func (i *collectionSimpleIndex) Save(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	key, err := i.getDocKey(doc)
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
	key, err := i.getDocKey(oldDoc)
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

func (i *collectionSimpleIndex) Name() string {
	return i.desc.Name
}

func (i *collectionSimpleIndex) Description() client.IndexDescription {
	return i.desc
}
