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

func getFieldValConverter(kind client.FieldKind) func(any) ([]byte, error) {
	switch kind {
	case client.FieldKind_STRING:
		return func(val any) ([]byte, error) {
			return []byte(val.(string)), nil
		}
	case client.FieldKind_INT:
		return func(val any) ([]byte, error) {
			intVal, ok := val.(int64)
			if !ok {
				return nil, errors.New("invalid int value")
			}
			return []byte(strconv.FormatInt(intVal, 10)), nil
		}
	case client.FieldKind_FLOAT:
		return func(val any) ([]byte, error) {
			floatVal, ok := val.(float64)
			if !ok {
				return nil, errors.New("invalid float value")
			}
			return []byte(strconv.FormatFloat(floatVal, 'f', -1, 64)), nil
		}
	case client.FieldKind_BOOL:
		return func(val any) ([]byte, error) {
			boolVal, ok := val.(bool)
			if !ok {
				return nil, errors.New("invalid bool value")
			}
			var intVal int64 = 0
			if boolVal {
				intVal = 1
			}
			return []byte(strconv.FormatInt(intVal, 10)), nil
		}
	case client.FieldKind_DATETIME:
		return func(val any) ([]byte, error) {
			timeStrVal := val.(string)
			_, err := time.Parse(time.RFC3339, timeStrVal)
			if err != nil {
				return nil, err
			}
			return []byte(timeStrVal), nil
		}
	default:
		panic("there is no test for this case")
	}
}

// NewCollectionIndex creates a new collection index
func NewCollectionIndex(
	collection client.Collection,
	desc client.IndexDescription,
) CollectionIndex {
	index := &collectionSimpleIndex{collection: collection, desc: desc}
	schema := collection.Description().Schema
	fieldID := schema.GetFieldKey(desc.Fields[0].Name)
	field := schema.Fields[fieldID]
	index.convertFunc = getFieldValConverter(field.Kind)
	return index
}

type collectionSimpleIndex struct {
	collection  client.Collection
	desc        client.IndexDescription
	convertFunc func(any) ([]byte, error)
}

var _ CollectionIndex = (*collectionSimpleIndex)(nil)

func (i *collectionSimpleIndex) getDocKey(doc *client.Document) (core.IndexDataStoreKey, error) {
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
			return core.IndexDataStoreKey{}, NewErrCanNotIndexInvalidFieldValue(err)
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
		field, _ := i.collection.Description().GetFieldByID(strconv.Itoa(int(key.IndexID)))
		return NewErrFailedToStoreIndexedField(field.Name, err)
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

func iteratePrefixKeys(
	ctx context.Context,
	prefix string,
	storage ds.Read,
	execFunc func(context.Context, ds.Key) error,
) error {
	q, err := storage.Query(ctx, query.Query{Prefix: prefix})
	if err != nil {
		return err
	}

	for res := range q.Next() {
		if res.Error != nil {
			_ = q.Close()
			return res.Error
		}
		err = execFunc(ctx, ds.NewKey(res.Key))
		if err != nil {
			_ = q.Close()
			return err
		}
	}
	if err = q.Close(); err != nil {
		return err
	}

	return nil
}
func (i *collectionSimpleIndex) RemoveAll(ctx context.Context, txn datastore.Txn) error {
	prefixKey := core.IndexDataStoreKey{}
	prefixKey.CollectionID = i.collection.ID()
	prefixKey.IndexID = i.desc.ID

	err := iteratePrefixKeys(ctx, prefixKey.ToString(), txn.Datastore(),
		func(ctx context.Context, key ds.Key) error {
			err := txn.Datastore().Delete(ctx, key)
			if err != nil {
				return NewCanNotDeleteIndexedField(err)
			}
			return nil
		})

	return err
}

func (i *collectionSimpleIndex) Name() string {
	return i.desc.Name
}

func (i *collectionSimpleIndex) Description() client.IndexDescription {
	return i.desc
}
