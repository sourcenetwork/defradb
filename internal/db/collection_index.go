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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema"
)

// createCollectionIndex creates a new collection index and saves it to the database in its system store.
func (db *db) createCollectionIndex(
	ctx context.Context,
	collectionName string,
	desc client.IndexDescription,
) (client.IndexDescription, error) {
	col, err := db.getCollectionByName(ctx, collectionName)
	if err != nil {
		return client.IndexDescription{}, NewErrCanNotReadCollection(collectionName, err)
	}
	return col.CreateIndex(ctx, desc)
}

func (db *db) dropCollectionIndex(
	ctx context.Context,
	collectionName, indexName string,
) error {
	col, err := db.getCollectionByName(ctx, collectionName)
	if err != nil {
		return NewErrCanNotReadCollection(collectionName, err)
	}
	return col.DropIndex(ctx, indexName)
}

// getAllIndexDescriptions returns all the index descriptions in the database.
func (db *db) getAllIndexDescriptions(
	ctx context.Context,
) (map[client.CollectionName][]client.IndexDescription, error) {
	// callers of this function must set a context transaction
	txn := mustGetContextTxn(ctx)
	prefix := core.NewCollectionIndexKey(immutable.None[uint32](), "")

	keys, indexDescriptions, err := datastore.DeserializePrefix[client.IndexDescription](ctx,
		prefix.ToString(), txn.Systemstore())

	if err != nil {
		return nil, err
	}

	indexes := make(map[client.CollectionName][]client.IndexDescription)

	for i := range keys {
		indexKey, err := core.NewCollectionIndexKeyFromString(keys[i])
		if err != nil {
			return nil, NewErrInvalidStoredIndexKey(indexKey.ToString())
		}

		col, err := description.GetCollectionByID(ctx, txn, indexKey.CollectionID.Value())
		if err != nil {
			return nil, err
		}

		indexes[col.Name.Value()] = append(
			indexes[col.Name.Value()],
			indexDescriptions[i],
		)
	}

	return indexes, nil
}

func (db *db) fetchCollectionIndexDescriptions(
	ctx context.Context,
	colID uint32,
) ([]client.IndexDescription, error) {
	// callers of this function must set a context transaction
	txn := mustGetContextTxn(ctx)
	prefix := core.NewCollectionIndexKey(immutable.Some(colID), "")
	_, indexDescriptions, err := datastore.DeserializePrefix[client.IndexDescription](
		ctx,
		prefix.ToString(),
		txn.Systemstore(),
	)
	if err != nil {
		return nil, err
	}
	return indexDescriptions, nil
}

func (c *collection) updateDocIndex(ctx context.Context, oldDoc, newDoc *client.Document) error {
	err := c.deleteIndexedDoc(ctx, oldDoc)
	if err != nil {
		return err
	}

	return c.indexNewDoc(ctx, newDoc)
}

func (c *collection) indexNewDoc(ctx context.Context, doc *client.Document) error {
	err := c.loadIndexes(ctx)
	if err != nil {
		return err
	}
	// callers of this function must set a context transaction
	txn := mustGetContextTxn(ctx)
	for _, index := range c.indexes {
		err = index.Save(ctx, txn, doc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *collection) updateIndexedDoc(
	ctx context.Context,
	doc *client.Document,
) error {
	err := c.loadIndexes(ctx)
	if err != nil {
		return err
	}
	// TODO-ACP: https://github.com/sourcenetwork/defradb/issues/2365 - ACP <> Indexing, possibly also check
	// and handle the case of when oldDoc == nil (will be nil if inaccessible document).
	oldDoc, err := c.get(
		ctx,
		c.getPrimaryKeyFromDocID(doc.ID()),
		c.Definition().CollectIndexedFields(),
		false,
	)
	if err != nil {
		return err
	}
	txn := mustGetContextTxn(ctx)
	for _, index := range c.indexes {
		err = index.Update(ctx, txn, oldDoc, doc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *collection) deleteIndexedDoc(
	ctx context.Context,
	doc *client.Document,
) error {
	err := c.loadIndexes(ctx)
	if err != nil {
		return err
	}
	txn := mustGetContextTxn(ctx)
	for _, index := range c.indexes {
		err = index.Delete(ctx, txn, doc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *collection) deleteIndexedDocWithID(
	ctx context.Context,
	docID client.DocID,
) error {
	doc, err := c.get(
		ctx,
		c.getPrimaryKeyFromDocID(docID),
		c.Definition().CollectIndexedFields(),
		false,
	)
	if err != nil {
		return err
	}
	return c.deleteIndexedDoc(ctx, doc)
}

// CreateIndex creates a new index on the collection.
//
// If the index name is empty, a name will be automatically generated.
// Otherwise its uniqueness will be checked against existing indexes and
// it will be validated with `schema.IsValidIndexName` method.
//
// The provided index description must include at least one field with
// a name that exists in the collection schema.
// Also it's `ID` field must be zero. It will be assigned a unique
// incremental value by the database.
//
// The index description will be stored in the system store.
//
// Once finished, if there are existing documents in the collection,
// the documents will be indexed by the new index.
func (c *collection) CreateIndex(
	ctx context.Context,
	desc client.IndexDescription,
) (client.IndexDescription, error) {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return client.IndexDescription{}, err
	}
	defer txn.Discard(ctx)

	index, err := c.createIndex(ctx, desc)
	if err != nil {
		return client.IndexDescription{}, err
	}
	return index.Description(), txn.Commit(ctx)
}

func (c *collection) createIndex(
	ctx context.Context,
	desc client.IndexDescription,
) (CollectionIndex, error) {
	if desc.Name != "" && !schema.IsValidIndexName(desc.Name) {
		return nil, schema.NewErrIndexWithInvalidName("!")
	}
	err := validateIndexDescription(desc)
	if err != nil {
		return nil, err
	}

	err = c.checkExistingFieldsAndAdjustRelFieldNames(desc.Fields)
	if err != nil {
		return nil, err
	}

	indexKey, err := c.generateIndexNameIfNeededAndCreateKey(ctx, &desc)
	if err != nil {
		return nil, err
	}

	colSeq, err := c.db.getSequence(
		ctx,
		core.NewIndexIDSequenceKey(c.ID()),
	)
	if err != nil {
		return nil, err
	}
	colID, err := colSeq.next(ctx)
	if err != nil {
		return nil, err
	}
	desc.ID = uint32(colID)

	buf, err := json.Marshal(desc)
	if err != nil {
		return nil, err
	}

	txn := mustGetContextTxn(ctx)
	err = txn.Systemstore().Put(ctx, indexKey.ToDS(), buf)
	if err != nil {
		return nil, err
	}
	colIndex, err := NewCollectionIndex(c, desc)
	if err != nil {
		return nil, err
	}
	c.def.Description.Indexes = append(c.def.Description.Indexes, colIndex.Description())
	c.indexes = append(c.indexes, colIndex)
	err = c.indexExistingDocs(ctx, colIndex)
	if err != nil {
		removeErr := colIndex.RemoveAll(ctx, txn)
		return nil, errors.Join(err, removeErr)
	}
	return colIndex, nil
}

func (c *collection) iterateAllDocs(
	ctx context.Context,
	fields []client.FieldDefinition,
	exec func(doc *client.Document) error,
) error {
	txn := mustGetContextTxn(ctx)
	identity := GetContextIdentity(ctx)

	df := c.newFetcher()
	err := df.Init(
		ctx,
		identity,
		txn,
		c.db.acp,
		c,
		fields,
		nil,
		nil,
		false,
		false,
	)
	if err != nil {
		return errors.Join(err, df.Close())
	}
	start := base.MakeDataStoreKeyWithCollectionDescription(c.Description())
	spans := core.NewSpans(core.NewSpan(start, start.PrefixEnd()))

	err = df.Start(ctx, spans)
	if err != nil {
		return errors.Join(err, df.Close())
	}

	for {
		encodedDoc, _, err := df.FetchNext(ctx)
		if err != nil {
			return errors.Join(err, df.Close())
		}
		if encodedDoc == nil {
			break
		}

		doc, err := fetcher.Decode(encodedDoc, c.Definition())
		if err != nil {
			return errors.Join(err, df.Close())
		}

		err = exec(doc)
		if err != nil {
			return errors.Join(err, df.Close())
		}
	}

	return df.Close()
}

func (c *collection) indexExistingDocs(
	ctx context.Context,
	index CollectionIndex,
) error {
	fields := make([]client.FieldDefinition, 0, 1)
	for _, field := range index.Description().Fields {
		colField, ok := c.Definition().GetFieldByName(field.Name)
		if ok {
			fields = append(fields, colField)
		}
	}
	txn := mustGetContextTxn(ctx)
	return c.iterateAllDocs(ctx, fields, func(doc *client.Document) error {
		return index.Save(ctx, txn, doc)
	})
}

// DropIndex removes an index from the collection.
//
// The index will be removed from the system store.
//
// All index artifacts for existing documents related the index will be removed.
func (c *collection) DropIndex(ctx context.Context, indexName string) error {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}
	defer txn.Discard(ctx)

	err = c.dropIndex(ctx, indexName)
	if err != nil {
		return err
	}
	return txn.Commit(ctx)
}

func (c *collection) dropIndex(ctx context.Context, indexName string) error {
	err := c.loadIndexes(ctx)
	if err != nil {
		return err
	}
	txn := mustGetContextTxn(ctx)

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

	for i := range c.Description().Indexes {
		if c.Description().Indexes[i].Name == indexName {
			c.def.Description.Indexes = append(c.Description().Indexes[:i], c.Description().Indexes[i+1:]...)
			break
		}
	}
	key := core.NewCollectionIndexKey(immutable.Some(c.ID()), indexName)
	err = txn.Systemstore().Delete(ctx, key.ToDS())
	if err != nil {
		return err
	}

	return nil
}

func (c *collection) dropAllIndexes(ctx context.Context) error {
	// callers of this function must set a context transaction
	txn := mustGetContextTxn(ctx)
	prefix := core.NewCollectionIndexKey(immutable.Some(c.ID()), "")

	keys, err := datastore.FetchKeysForPrefix(ctx, prefix.ToString(), txn.Systemstore())
	if err != nil {
		return err
	}

	for _, key := range keys {
		err = txn.Systemstore().Delete(ctx, key)
		if err != nil {
			return err
		}
	}

	return err
}

func (c *collection) loadIndexes(ctx context.Context) error {
	indexDescriptions, err := c.db.fetchCollectionIndexDescriptions(ctx, c.ID())
	if err != nil {
		return err
	}
	colIndexes := make([]CollectionIndex, 0, len(indexDescriptions))
	for _, indexDesc := range indexDescriptions {
		index, err := NewCollectionIndex(c, indexDesc)
		if err != nil {
			return err
		}
		colIndexes = append(colIndexes, index)
	}
	c.def.Description.Indexes = indexDescriptions
	c.indexes = colIndexes
	return nil
}

// GetIndexes returns all indexes for the collection.
func (c *collection) GetIndexes(ctx context.Context) ([]client.IndexDescription, error) {
	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return nil, err
	}
	defer txn.Discard(ctx)

	err = c.loadIndexes(ctx)
	if err != nil {
		return nil, err
	}
	return c.Description().Indexes, nil
}

// checkExistingFieldsAndAdjustRelFieldNames checks if the fields in the index description
// exist in the collection schema.
// If a field is a relation, it will be adjusted to relation id field name, a.k.a. `field_name + _id`.
func (c *collection) checkExistingFieldsAndAdjustRelFieldNames(
	fields []client.IndexedFieldDescription,
) error {
	for i := range fields {
		field, found := c.Schema().GetFieldByName(fields[i].Name)
		if !found {
			return NewErrNonExistingFieldForIndex(fields[i].Name)
		}
		if field.Kind.IsObject() {
			fields[i].Name = fields[i].Name + request.RelatedObjectID
		}
	}
	return nil
}

func (c *collection) generateIndexNameIfNeededAndCreateKey(
	ctx context.Context,
	desc *client.IndexDescription,
) (core.CollectionIndexKey, error) {
	// callers of this function must set a context transaction
	txn := mustGetContextTxn(ctx)

	var indexKey core.CollectionIndexKey
	if desc.Name == "" {
		nameIncrement := 1
		for {
			desc.Name = generateIndexName(c, desc.Fields, nameIncrement)
			indexKey = core.NewCollectionIndexKey(immutable.Some(c.ID()), desc.Name)
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
		indexKey = core.NewCollectionIndexKey(immutable.Some(c.ID()), desc.Name)
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
	for i := range desc.Fields {
		if desc.Fields[i].Name == "" {
			return ErrIndexFieldMissingName
		}
	}
	return nil
}

func generateIndexName(col client.Collection, fields []client.IndexedFieldDescription, inc int) string {
	sb := strings.Builder{}
	// at the moment we support only single field indexes that can be stored only in
	// ascending order. This will change once we introduce composite indexes.
	direction := "ASC"
	if col.Name().HasValue() {
		sb.WriteString(col.Name().Value())
	} else {
		sb.WriteString(fmt.Sprint(col.ID()))
	}
	sb.WriteByte('_')
	// we can safely assume that there is at least one field in the slice
	// because we validate it before calling this function
	sb.WriteString(fields[0].Name)
	sb.WriteByte('_')
	sb.WriteString(direction)
	if inc > 1 {
		sb.WriteByte('_')
		sb.WriteString(strconv.Itoa(inc))
	}
	return sb.String()
}
