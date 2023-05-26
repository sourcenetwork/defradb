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
	"encoding/json"

	"github.com/ipfs/go-datastore/query"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
)

// createCollectionIndex creates a new collection index and saves it to the database in its system store.
func (db *db) createCollectionIndex(
	ctx context.Context,
	txn datastore.Txn,
	collectionName string,
	desc client.IndexDescription,
) (client.IndexDescription, error) {
	col, err := db.getCollectionByName(ctx, txn, collectionName)
	if err != nil {
		return client.IndexDescription{}, NewErrCollectionDoesntExist(collectionName)
	}
	col = col.WithTxn(txn)
	return col.CreateIndex(ctx, desc)
}

func (db *db) dropCollectionIndex(
	ctx context.Context,
	txn datastore.Txn,
	collectionName, indexName string,
) error {
	col, err := db.getCollectionByName(ctx, txn, collectionName)
	if err != nil {
		return NewErrCollectionDoesntExist(collectionName)
	}
	col = col.WithTxn(txn)
	return col.DropIndex(ctx, indexName)
}

// getAllCollectionIndexes returns all the indexes in the database.
func (db *db) getAllCollectionIndexes(
	ctx context.Context,
	txn datastore.Txn,
) ([]client.CollectionIndexDescription, error) {
	prefix := core.NewCollectionIndexKey("", "")
	q, err := txn.Systemstore().Query(ctx, query.Query{
		Prefix: prefix.ToString(),
	})
	if err != nil {
		return nil, NewErrFailedToCreateCollectionQuery(err)
	}
	defer func() {
		if err := q.Close(); err != nil {
			log.ErrorE(ctx, "Failed to close collection query", err)
		}
	}()

	indexes := make([]client.CollectionIndexDescription, 0)
	for res := range q.Next() {
		if res.Error != nil {
			return nil, err
		}

		var colDesk client.IndexDescription
		err = json.Unmarshal(res.Value, &colDesk)
		if err != nil {
			return nil, NewErrInvalidStoredIndex(err)
		}
		indexKey, err := core.NewCollectionIndexKeyFromString(res.Key)
		if err != nil {
			return nil, NewErrInvalidStoredIndexKey(indexKey.ToString())
		}
		indexes = append(indexes, client.CollectionIndexDescription{
			CollectionName: indexKey.CollectionName,
			Index:          colDesk,
		})
	}

	return indexes, nil
}

func (db *db) getCollectionIndexes(
	ctx context.Context,
	txn datastore.Txn,
	colName string,
) ([]client.IndexDescription, error) {
	prefix := core.NewCollectionIndexKey(colName, "")
	q, err := txn.Systemstore().Query(ctx, query.Query{
		Prefix: prefix.ToString(),
	})
	if err != nil {
		return nil, NewErrFailedToCreateCollectionQuery(err)
	}
	defer func() {
		if err := q.Close(); err != nil {
			log.ErrorE(ctx, "Failed to close collection query", err)
		}
	}()

	indexes := make([]client.IndexDescription, 0)
	for res := range q.Next() {
		if res.Error != nil {
			return nil, err
		}

		var colDesk client.IndexDescription
		err = json.Unmarshal(res.Value, &colDesk)
		if err != nil {
			return nil, NewErrInvalidStoredIndex(err)
		}
		indexes = append(indexes, colDesk)
	}

	return indexes, nil
}

func (c *collection) indexNewDoc(ctx context.Context, txn datastore.Txn, doc *client.Document) error {
	indexes, err := c.getIndexes(ctx, txn)
	if err != nil {
		return err
	}
	for _, index := range indexes {
		err = index.Save(ctx, txn, doc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *collection) updateIndex(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	f := new(fetcher.DocumentFetcher)
	fields := make([]*client.FieldDescription, len(c.desc.Schema.Fields))
	for i, field := range c.desc.Schema.Fields {
		fields[i] = &field
	}
	err := f.Init(&c.desc, fields, false, false)
	err = err

	docKey := base.MakeDocKey(c.Description(), doc.Key().String())
	err = f.Start(ctx, txn, core.NewSpans(core.NewSpan(docKey, docKey.PrefixEnd())))
	oldDoc, err := f.FetchNextDecoded(ctx)
	_, err = c.getIndexes(ctx, txn)
	for _, index := range c.indexes {
		err = index.Update(ctx, txn, oldDoc, doc)
	}
	err = f.Close()
	return nil
}
