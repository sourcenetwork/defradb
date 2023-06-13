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
	"fmt"
	"strconv"
	"strings"

	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/request/graphql/schema"
)

// collectionIndexDescription describes an index on a collection.
// It's useful for retrieving a list of indexes without having to
// retrieve the entire collection description.
type collectionIndexDescription struct {
	// CollectionName contains the name of the collection.
	CollectionName string
	// Index contains the index description.
	Index client.IndexDescription
}

// createCollectionIndex creates a new collection index and saves it to the database in its system store.
func (db *db) createCollectionIndex(
	ctx context.Context,
	txn datastore.Txn,
	collectionName string,
	desc client.IndexDescription,
) (client.IndexDescription, error) {
	col, err := db.getCollectionByName(ctx, txn, collectionName)
	if err != nil {
		return client.IndexDescription{}, NewErrCanNotReadCollection(collectionName, err)
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
		return NewErrCanNotReadCollection(collectionName, err)
	}
	col = col.WithTxn(txn)
	return col.DropIndex(ctx, indexName)
}

// getAllCollectionIndexes returns all the indexes in the database.
func (db *db) getAllCollectionIndexes(
	ctx context.Context,
	txn datastore.Txn,
) ([]collectionIndexDescription, error) {
	prefix := core.NewCollectionIndexKey("", "")

	indexMap, err := deserializePrefix[client.IndexDescription](ctx,
		prefix.ToString(), txn.Systemstore())

	if err != nil {
		return nil, err
	}

	indexes := make([]collectionIndexDescription, 0, len(indexMap))

	for indexKeyStr, index := range indexMap {
		indexKey, err := core.NewCollectionIndexKeyFromString(indexKeyStr)
		if err != nil {
			return nil, NewErrInvalidStoredIndexKey(indexKey.ToString())
		}
		indexes = append(indexes, collectionIndexDescription{
			CollectionName: indexKey.CollectionName,
			Index:          index,
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
	indexMap, err := deserializePrefix[client.IndexDescription](ctx,
		prefix.ToString(), txn.Systemstore())
	if err != nil {
		return nil, err
	}
	indexes := make([]client.IndexDescription, 0, len(indexMap))
	for _, index := range indexMap {
		indexes = append(indexes, index)
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

func (c *collection) collectIndexedFields() []*client.FieldDescription {
	fieldsMap := make(map[string]*client.FieldDescription)
	for _, index := range c.indexes {
		for _, field := range index.Description().Fields {
			for i := range c.desc.Schema.Fields {
				colField := &c.desc.Schema.Fields[i]
				if field.Name == colField.Name {
					fieldsMap[field.Name] = colField
					break
				}
			}
		}
	}
	fields := make([]*client.FieldDescription, 0, len(fieldsMap))
	for _, field := range fieldsMap {
		fields = append(fields, field)
	}
	return fields
}

func (c *collection) updateIndex(
	ctx context.Context,
	txn datastore.Txn,
	doc *client.Document,
) error {
	_, err := c.getIndexes(ctx, txn)
	if err != nil {
		return err
	}
	oldDoc, err := c.get(ctx, txn, c.getPrimaryKeyFromDocKey(doc.Key()), c.collectIndexedFields(), false)
	if err != nil {
		return err
	}
	for _, index := range c.indexes {
		err = index.Update(ctx, txn, oldDoc, doc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *collection) CreateIndex(
	ctx context.Context,
	desc client.IndexDescription,
) (client.IndexDescription, error) {
	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return client.IndexDescription{}, err
	}

	index, err := c.createIndex(ctx, txn, desc)
	if err != nil {
		return client.IndexDescription{}, err
	}
	if c.isIndexCached {
		c.indexes = append(c.indexes, index)
	}
	err = c.indexExistingDocs(ctx, txn, index)
	if err != nil {
		return client.IndexDescription{}, err
	}
	return index.Description(), nil
}

func (c *collection) newFetcher() fetcher.Fetcher {
	if c.fetcherFactory != nil {
		return c.fetcherFactory()
	} else {
		return new(fetcher.DocumentFetcher)
	}
}

func (c *collection) iterateAllDocs(
	ctx context.Context,
	txn datastore.Txn,
	fields []*client.FieldDescription,
	exec func(doc *client.Document) error,
) error {
	df := c.newFetcher()
	err := df.Init(&c.desc, fields, false, false)
	if err != nil {
		_ = df.Close()
		return err
	}
	start := base.MakeCollectionKey(c.desc)
	spans := core.NewSpans(core.NewSpan(start, start.PrefixEnd()))

	err = df.Start(ctx, txn, spans)
	if err != nil {
		_ = df.Close()
		return err
	}

	var doc *client.Document
	for {
		doc, err = df.FetchNextDecoded(ctx)
		if err != nil {
			_ = df.Close()
			return err
		}
		if doc == nil {
			break
		}
		err = exec(doc)
		if err != nil {
			return err
		}
	}

	return df.Close()
}

func (c *collection) indexExistingDocs(
	ctx context.Context,
	txn datastore.Txn,
	index CollectionIndex,
) error {
	fields := make([]*client.FieldDescription, 0, 1)
	for _, field := range index.Description().Fields {
		for i := range c.desc.Schema.Fields {
			colField := &c.desc.Schema.Fields[i]
			if field.Name == colField.Name {
				fields = append(fields, colField)
				break
			}
		}
	}

	return c.iterateAllDocs(ctx, txn, fields, func(doc *client.Document) error {
		return index.Save(ctx, txn, doc)
	})
}

func (c *collection) DropIndex(ctx context.Context, indexName string) error {
	key := core.NewCollectionIndexKey(c.Name(), indexName)

	txn, err := c.getTxn(ctx, false)
	if err != nil {
		return err
	}
	_, err = c.getIndexes(ctx, txn)
	if err != nil {
		return err
	}
	var didFind bool
	for i := range c.indexes {
		if c.indexes[i].Name() == indexName {
			err = c.indexes[i].RemoveAll(ctx, txn)
			if err != nil {
				return err
			}
			c.indexes = append(c.indexes[:i], c.indexes[i+1:]...)
			didFind = true
			break
		}
	}
	if !didFind {
		return NewErrIndexWithNameDoesNotExists(indexName)
	}

	for i := range c.desc.Indexes {
		if c.desc.Indexes[i].Name == indexName {
			c.desc.Indexes = append(c.desc.Indexes[:i], c.desc.Indexes[i+1:]...)
			break
		}
	}
	err = txn.Systemstore().Delete(ctx, key.ToDS())
	if err != nil {
		return err
	}

	return nil
}

func (c *collection) dropAllIndexes(ctx context.Context, txn datastore.Txn) error {
	prefix := core.NewCollectionIndexKey(c.Name(), "")

	err := iteratePrefixKeys(ctx, prefix.ToString(), txn.Systemstore(),
		func(ctx context.Context, key ds.Key) error {
			return txn.Systemstore().Delete(ctx, key)
		})

	return err
}

func (c *collection) getIndexes(ctx context.Context, txn datastore.Txn) ([]CollectionIndex, error) {
	if c.isIndexCached {
		return c.indexes, nil
	}

	prefix := core.NewCollectionIndexKey(c.Name(), "")
	indexes, err := deserializePrefix[client.IndexDescription](ctx, prefix.ToString(), txn.Systemstore())
	if err != nil {
		return nil, err
	}
	colIndexes := make([]CollectionIndex, 0, len(indexes))
	for _, index := range indexes {
		colIndexes = append(colIndexes, NewCollectionIndex(c, index))
	}

	descriptions := make([]client.IndexDescription, 0, len(colIndexes))
	for _, index := range colIndexes {
		descriptions = append(descriptions, index.Description())
	}
	c.desc.Indexes = descriptions
	c.indexes = colIndexes
	c.isIndexCached = true
	return colIndexes, nil
}

func (c *collection) GetIndexes(ctx context.Context) ([]client.IndexDescription, error) {
	txn, err := c.getTxn(ctx, true)
	if err != nil {
		return nil, err
	}
	indexes, err := c.getIndexes(ctx, txn)
	if err != nil {
		return nil, err
	}
	indexDescriptions := make([]client.IndexDescription, 0, len(indexes))
	for _, index := range indexes {
		indexDescriptions = append(indexDescriptions, index.Description())
	}

	return indexDescriptions, nil
}

func (c *collection) createIndex(
	ctx context.Context,
	txn datastore.Txn,
	desc client.IndexDescription,
) (CollectionIndex, error) {
	if desc.Name != "" && !schema.IsValidIndexName(desc.Name) {
		return nil, schema.NewErrIndexWithInvalidName("!")
	}
	err := validateIndexDescription(desc)
	if err != nil {
		return nil, err
	}

	err = c.checkExistingFields(ctx, desc.Fields)
	if err != nil {
		return nil, err
	}

	indexKey, err := c.processIndexName(ctx, txn, &desc)
	if err != nil {
		return nil, err
	}

	colSeq, err := c.db.getSequence(ctx, txn, fmt.Sprintf("%s/%d", core.COLLECTION_INDEX, c.ID()))
	if err != nil {
		return nil, err
	}
	colID, err := colSeq.next(ctx, txn)
	if err != nil {
		return nil, err
	}
	desc.ID = uint32(colID)

	buf, err := json.Marshal(desc)
	if err != nil {
		return nil, err
	}

	err = txn.Systemstore().Put(ctx, indexKey.ToDS(), buf)
	if err != nil {
		return nil, err
	}
	colIndex := NewCollectionIndex(c, desc)
	c.desc.Indexes = append(c.desc.Indexes, colIndex.Description())
	return colIndex, nil
}

func (c *collection) checkExistingFields(
	ctx context.Context,
	fields []client.IndexedFieldDescription,
) error {
	collectionFields := c.Description().Schema.Fields
	for _, field := range fields {
		found := false
		fieldLower := strings.ToLower(field.Name)
		for _, colField := range collectionFields {
			if fieldLower == strings.ToLower(colField.Name) {
				found = true
				break
			}
		}
		if !found {
			return NewErrNonExistingFieldForIndex(field.Name)
		}
	}
	return nil
}

func (c *collection) processIndexName(
	ctx context.Context,
	txn datastore.Txn,
	desc *client.IndexDescription,
) (core.CollectionIndexKey, error) {
	var indexKey core.CollectionIndexKey
	if desc.Name == "" {
		nameIncrement := 1
		for {
			desc.Name = generateIndexName(c, desc.Fields, nameIncrement)
			indexKey = core.NewCollectionIndexKey(c.Name(), desc.Name)
			exists, err := txn.Systemstore().Has(ctx, indexKey.ToDS())
			if err != nil {
				return core.CollectionIndexKey{}, err
			}
			if !exists {
				break
			}
			nameIncrement++
		}
	} else {
		indexKey = core.NewCollectionIndexKey(c.Name(), desc.Name)
		exists, err := txn.Systemstore().Has(ctx, indexKey.ToDS())
		if err != nil {
			return core.CollectionIndexKey{}, err
		}
		if exists {
			return core.CollectionIndexKey{}, NewErrIndexWithNameAlreadyExists(desc.Name)
		}
	}
	return indexKey, nil
}

func validateIndexDescription(desc client.IndexDescription) error {
	if desc.ID != 0 {
		return NewErrNonZeroIndexIDProvided(desc.ID)
	}
	if len(desc.Fields) == 0 {
		return ErrIndexMissingFields
	}
	if len(desc.Fields) == 1 && desc.Fields[0].Direction == client.Descending {
		return ErrIndexSingleFieldWrongDirection
	}
	for i := range desc.Fields {
		if desc.Fields[i].Name == "" {
			return ErrIndexFieldMissingName
		}
		if desc.Fields[i].Direction == "" {
			desc.Fields[i].Direction = client.Ascending
		}
	}
	return nil
}

func generateIndexName(col client.Collection, fields []client.IndexedFieldDescription, inc int) string {
	sb := strings.Builder{}
	direction := "ASC"
	sb.WriteString(col.Name())
	sb.WriteByte('_')
	sb.WriteString(fields[0].Name)
	sb.WriteByte('_')
	sb.WriteString(direction)
	if inc > 1 {
		sb.WriteByte('_')
		sb.WriteString(strconv.Itoa(inc))
	}
	return sb.String()
}

func deserializePrefix[T any](ctx context.Context, prefix string, storage ds.Read) (map[string]T, error) {
	q, err := storage.Query(ctx, query.Query{Prefix: prefix})
	if err != nil {
		return nil, NewErrFailedToCreateCollectionQuery(err)
	}
	defer func() {
		if err := q.Close(); err != nil {
			log.ErrorE(ctx, "Failed to close collection query", err)
		}
	}()

	elements := make(map[string]T)
	for res := range q.Next() {
		if res.Error != nil {
			return nil, res.Error
		}

		var element T
		err = json.Unmarshal(res.Value, &element)
		if err != nil {
			return nil, NewErrInvalidStoredIndex(err)
		}
		elements[res.Key] = element
	}
	return elements, nil
}
