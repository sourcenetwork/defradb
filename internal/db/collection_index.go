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
	"strconv"
	"strings"

	"github.com/sourcenetwork/immutable"

	"slices"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/db/sequence"
	"github.com/sourcenetwork/defradb/internal/db/txnctx"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema"
)

// getAllIndexDescriptions returns all the index descriptions in the database.
func (db *DB) getAllIndexDescriptions(
	ctx context.Context,
) (map[client.CollectionName][]client.IndexDescription, error) {
	txn := txnctx.MustGet(ctx)
	collections, err := description.GetCollections(ctx, txn)

	if err != nil {
		return nil, err
	}

	indexes := make(map[client.CollectionName][]client.IndexDescription)

	for _, col := range collections {
		if len(col.Indexes) > 0 {
			indexes[col.Name] = col.Indexes
		}
	}

	return indexes, nil
}

func (c *collection) updateDocIndex(ctx context.Context, oldDoc, newDoc *client.Document) error {
	err := c.deleteIndexedDoc(ctx, oldDoc)
	if err != nil {
		return err
	}

	return c.indexNewDoc(ctx, newDoc)
}

func (c *collection) indexNewDoc(ctx context.Context, doc *client.Document) error {
	// callers of this function must set a context transaction
	txn := txnctx.MustGet(ctx)
	for _, index := range c.indexes {
		err := index.Save(ctx, txn, doc)
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
	primaryKey, err := c.getPrimaryKeyFromDocID(ctx, doc.ID())
	if err != nil {
		return err
	}

	// TODO-ACP: https://github.com/sourcenetwork/defradb/issues/2365 - ACP <> Indexing, possibly also check
	// and handle the case of when oldDoc == nil (will be nil if inaccessible document).
	oldDoc, err := c.get(
		ctx,
		primaryKey,
		c.Definition().CollectIndexedFields(),
		false,
	)
	if err != nil {
		return err
	}
	txn := txnctx.MustGet(ctx)
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
	txn := txnctx.MustGet(ctx)
	for _, index := range c.indexes {
		err := index.Delete(ctx, txn, doc)
		if err != nil {
			return err
		}
	}
	return nil
}

// deleteIndexedDocWithID deletes an indexed document with the provided document ID.
func (c *collection) deleteIndexedDocWithID(
	ctx context.Context,
	docID client.DocID,
) error {
	primaryKey, err := c.getPrimaryKeyFromDocID(ctx, docID)
	if err != nil {
		return err
	}

	// we need to fetch the document to delete it from the indexes, because in order to do so
	// we need to know the values of the fields that are indexed.
	doc, err := c.get(
		ctx,
		primaryKey,
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
	desc client.IndexCreateRequest,
) (client.IndexDescription, error) {
	ctx, span := tracer.Start(ctx)
	defer span.End()

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

func processCreateIndexRequest(
	ctx context.Context,
	def client.CollectionDefinition,
	desc client.IndexCreateRequest,
) (client.IndexDescription, error) {
	err := validateIndexDescription(desc)
	if err != nil {
		return client.IndexDescription{}, err
	}

	err = checkExistingFieldsAndAdjustRelFieldNames(def.Schema, desc.Fields)
	if err != nil {
		return client.IndexDescription{}, err
	}

	indexName, err := generateIndexNameIfNeeded(def.Version, desc)
	if err != nil {
		return client.IndexDescription{}, err
	}

	txn := txnctx.MustGet(ctx)

	colSeq, err := sequence.Get(
		ctx,
		txn,
		keys.NewIndexIDSequenceKey(def.Version.CollectionID),
	)
	if err != nil {
		return client.IndexDescription{}, err
	}
	indexID, err := colSeq.Next(ctx, txn)
	if err != nil {
		return client.IndexDescription{}, err
	}

	return client.IndexDescription{
		Name:   indexName,
		ID:     uint32(indexID),
		Fields: desc.Fields,
		Unique: desc.Unique,
	}, nil
}

func (c *collection) createIndex(
	ctx context.Context,
	createReq client.IndexCreateRequest,
) (CollectionIndex, error) {
	desc, err := processCreateIndexRequest(ctx, c.Definition(), createReq)
	if err != nil {
		return nil, err
	}

	c.def.Version.Indexes = append(c.def.Version.Indexes, desc)

	txn := txnctx.MustGet(ctx)
	err = description.SaveCollection(ctx, txn, c.def.Version)
	if err != nil {
		c.def.Version.Indexes = c.def.Version.Indexes[:len(c.def.Version.Indexes)-1]
		return nil, err
	}

	index, err := c.addNewIndex(ctx, desc)
	if err != nil {
		c.def.Version.Indexes = c.def.Version.Indexes[:len(c.def.Version.Indexes)-1]
		return nil, err
	}

	return index, nil
}

func (c *collection) addNewIndex(ctx context.Context, desc client.IndexDescription) (CollectionIndex, error) {
	colIndex, err := NewCollectionIndex(c, desc)
	if err != nil {
		return nil, err
	}

	c.indexes = append(c.indexes, colIndex)

	err = c.indexExistingDocs(ctx, colIndex)
	if err != nil {
		txn := txnctx.MustGet(ctx)
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
	txn := txnctx.MustGet(ctx)

	df := c.newFetcher()
	err := df.Init(
		ctx,
		identity.FromContext(ctx),
		txn,
		c.db.documentACP,
		immutable.None[client.IndexDescription](),
		c,
		fields,
		nil,
		nil,
		nil,
		false,
	)
	if err != nil {
		return errors.Join(err, df.Close())
	}

	shortID, err := id.GetShortCollectionID(ctx, txn, c.Version().CollectionID)
	if err != nil {
		return err
	}

	prefix := keys.DataStoreKey{
		CollectionShortID: shortID,
	}
	err = df.Start(ctx, prefix)
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
	txn := txnctx.MustGet(ctx)
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
	ctx, span := tracer.Start(ctx)
	defer span.End()

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
	txn := txnctx.MustGet(ctx)

	var didFind bool
	for i := range c.indexes {
		if c.indexes[i].Name() == indexName {
			err := c.indexes[i].RemoveAll(ctx, txn)
			if err != nil {
				return err
			}
			c.indexes = slices.Delete(c.indexes, i, i+1)
			didFind = true
			break
		}
	}
	if !didFind {
		return NewErrIndexWithNameDoesNotExists(indexName)
	}

	for i := range c.Version().Indexes {
		if c.Version().Indexes[i].Name == indexName {
			c.def.Version.Indexes = slices.Delete(c.Version().Indexes, i, i+1)
			break
		}
	}

	return nil
}

// GetIndexes returns all indexes for the collection.
func (c *collection) GetIndexes(context.Context) ([]client.IndexDescription, error) {
	return c.Version().Indexes, nil
}

// checkExistingFieldsAndAdjustRelFieldNames checks if the fields in the index description
// exist in the collection schema.
// If a field is a relation, it will be adjusted to relation id field name, a.k.a. `field_name + _id`.
func checkExistingFieldsAndAdjustRelFieldNames(
	schema client.SchemaDescription,
	fields []client.IndexedFieldDescription,
) error {
	for i := range fields {
		field, found := schema.GetFieldByName(fields[i].Name)
		if !found {
			return NewErrNonExistingFieldForIndex(fields[i].Name)
		}
		if field.Kind.IsObject() {
			fields[i].Name = fields[i].Name + request.RelatedObjectID
		}
	}
	return nil
}

func generateIndexNameIfNeeded(
	colVersion client.CollectionVersion,
	createReq client.IndexCreateRequest,
) (string, error) {
	indexName := createReq.Name
	if indexName == "" {
		nameIncrement := 1
		for {
			indexName = generateIndexName(colVersion.Name, createReq.Fields, nameIncrement)

			isUnique := true
			for _, index := range colVersion.Indexes {
				if index.Name == indexName {
					isUnique = false
					break
				}
			}

			if isUnique {
				break
			}

			nameIncrement++
		}
	} else {
		for _, index := range colVersion.Indexes {
			if index.Name == indexName {
				return "", NewErrIndexWithNameAlreadyExists(indexName)
			}
		}
	}

	return indexName, nil
}

func validateIndexDescription(desc client.IndexCreateRequest) error {
	if desc.Name != "" && !schema.IsValidIndexName(desc.Name) {
		return schema.NewErrIndexWithInvalidName("!")
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

func generateIndexName(colName string, fields []client.IndexedFieldDescription, inc int) string {
	sb := strings.Builder{}
	// at the moment we support only single field indexes that can be stored only in
	// ascending order. This will change once we introduce composite indexes.
	direction := "ASC"
	sb.WriteString(colName)
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
