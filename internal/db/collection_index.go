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

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/db/sequence"
	"github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema"
	"github.com/sourcenetwork/defradb/internal/utils"
)

// listIndexDescriptions returns all the index descriptions in the database.
func (db *DB) listIndexDescriptions(
	ctx context.Context,
) (map[client.CollectionName][]client.IndexDescription, error) {
	collections, err := description.GetCollections(ctx)

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

	return c.addDocToIndex(ctx, newDoc)
}

func (c *collection) addDocToIndex(ctx context.Context, doc *client.Document) error {
	// callers of this function must set a context transaction
	for _, index := range c.indexes {
		err := index.Save(ctx, doc)
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
		c.Version().CollectIndexedFields(),
		false,
	)
	if err != nil {
		return err
	}
	for _, index := range c.indexes {
		err = index.Update(ctx, oldDoc, doc)
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
	for _, index := range c.indexes {
		err := index.Delete(ctx, doc)
		if err != nil {
			return NewErrDeleteIndexedDoc(err, index.Description().Name)
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
		c.Version().CollectIndexedFields(),
		false,
	)
	if err != nil {
		return err
	}
	if doc == nil {
		// If the document cannot be fetched (e.g., due to ACP restrictions),
		// skip index deletion. The caller (Delete) will handle the authorization
		// error in applyDelete.
		return nil
	}
	return c.deleteIndexedDoc(ctx, doc)
}

// NewIndex makes a new index on the collection.
//
// If the index name is empty, a name will be automatically generated.
// Otherwise its uniqueness will be checked against existing indexes and
// it will be validated with `schema.IsValidIndexName` method.
//
// The provided index description must include at least one field with
// a name that exists in the collection definition.
//
// The index description will be stored in the system store.
//
// Once finished, if there are existing documents in the collection,
// the documents will be indexed by the new index.
func (c *collection) NewIndex(
	ctx context.Context,
	desc client.NewIndexRequest,
	opts ...options.Enumerable[options.NewCollectionIndexOptions],
) (client.IndexDescription, error) {
	ctx, _, _ = getTxnAndSetCtxForCollection(ctx, c)

	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeNewIndexPerm); err != nil {
		return client.IndexDescription{}, err
	}

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return client.IndexDescription{}, err
	}

	defer txn.Discard()

	index, err := c.newIndex(ctx, desc)
	if err != nil {
		return client.IndexDescription{}, err
	}

	return index.Description(), txn.Commit()
}

func processNewIndexRequest(
	ctx context.Context,
	def client.CollectionVersion,
	desc client.NewIndexRequest,
) (client.IndexDescription, error) {
	err := validateIndexDescription(desc)
	if err != nil {
		return client.IndexDescription{}, err
	}

	err = checkExistingFieldsAndAdjustRelFieldNames(def, desc.Fields)
	if err != nil {
		return client.IndexDescription{}, err
	}

	indexName, err := generateIndexNameIfNeeded(def, desc)
	if err != nil {
		return client.IndexDescription{}, err
	}

	colSeq, err := sequence.Get(
		ctx,
		keys.NewIndexIDSequenceKey(def.CollectionID),
	)
	if err != nil {
		return client.IndexDescription{}, err
	}
	indexID, err := colSeq.Next(ctx)
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

func (c *collection) newIndex(
	ctx context.Context,
	newReq client.NewIndexRequest,
) (CollectionIndex, error) {
	desc, err := processNewIndexRequest(ctx, c.Version(), newReq)
	if err != nil {
		return nil, err
	}

	c.def.Indexes = append(c.def.Indexes, desc)

	err = description.SaveCollection(ctx, c.def)
	if err != nil {
		c.def.Indexes = c.def.Indexes[:len(c.def.Indexes)-1]
		return nil, err
	}

	index, err := c.appendNewIndexAndIndexExistingDocs(ctx, desc)
	if err != nil {
		c.def.Indexes = c.def.Indexes[:len(c.def.Indexes)-1]
		return nil, err
	}

	return index, nil
}

func (c *collection) appendNewIndexAndIndexExistingDocs(
	ctx context.Context,
	desc client.IndexDescription,
) (CollectionIndex, error) {
	colIndex, err := NewCollectionIndex(c, desc)
	if err != nil {
		return nil, err
	}

	c.indexes = append(c.indexes, colIndex)

	err = c.indexExistingDocs(ctx, colIndex)
	if err != nil {
		removeErr := colIndex.RemoveAll(ctx)
		return nil, errors.Join(err, removeErr)
	}

	return colIndex, nil
}

func (c *collection) iterateAllDocs(
	ctx context.Context,
	fields []client.CollectionFieldDescription,
	exec func(doc *client.Document) error,
) error {
	txn := datastore.CtxMustGetTxn(ctx)
	df := c.newFetcher(ctx)
	err := df.Init(
		ctx,
		identity.FromContext(ctx),
		txn,
		c.db.nodeACP,
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

	shortID, err := id.GetShortCollectionID(ctx, c.Version().CollectionID)
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

		doc, err := fetcher.Decode(ctx, encodedDoc, c.Version())
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
	fields := make([]client.CollectionFieldDescription, 0, len(index.Description().Fields))
	for _, field := range index.Description().Fields {
		colField, ok := c.Version().GetFieldByName(field.Name)
		if ok {
			fields = append(fields, colField)
		}
	}
	return c.iterateAllDocs(ctx, fields, func(doc *client.Document) error {
		return index.Save(ctx, doc)
	})
}

// DeleteIndex removes an index from the collection.
//
// The index will be removed from the system store.
//
// All index artifacts for existing documents related the index will be removed.
func (c *collection) DeleteIndex(
	ctx context.Context,
	indexName string,
	opts ...options.Enumerable[options.DeleteCollectionIndexOptions],
) error {
	ctx, _, _ = getTxnAndSetCtxForCollection(ctx, c)

	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeDeleteIndexPerm); err != nil {
		return err
	}

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}

	defer txn.Discard()

	err = c.deleteIndex(ctx, indexName)
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (c *collection) deleteIndex(ctx context.Context, indexName string) error {
	var didFind bool
	for i := range c.indexes {
		if c.indexes[i].Name() == indexName {
			err := c.indexes[i].RemoveAll(ctx)
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

	oldIndexes := make([]client.IndexDescription, len(c.Version().Indexes))
	copy(oldIndexes, c.Version().Indexes)
	for i := range c.Version().Indexes {
		if c.Version().Indexes[i].Name == indexName {
			c.def.Indexes = slices.Delete(c.Version().Indexes, i, i+1)
			break
		}
	}

	err := description.SaveCollection(ctx, c.def)
	if err != nil {
		c.def.Indexes = oldIndexes
		return err
	}

	return nil
}

// ListIndexes returns all indexes for the collection.
func (c *collection) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListCollectionIndexesOptions],
) ([]client.IndexDescription, error) {
	ctx, _, _ = getTxnAndSetCtxForCollection(ctx, c)

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeListIndexPerm); err != nil {
		return nil, err
	}

	return c.Version().Indexes, nil
}

// NewEncryptedIndex adds a new encrypted index to the collection.
func (c *collection) NewEncryptedIndex(
	ctx context.Context,
	addRequest client.EncryptedIndexDescription,
	opts ...options.Enumerable[options.NewEncryptedIndexOptions],
) (client.EncryptedIndexDescription, error) {
	ctx, _, _ = getTxnAndSetCtxForCollection(ctx, c)

	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)
	ident := opt.GetIdentity()

	if err := c.db.checkNodeAccess(ctx, ident, acpTypes.NodeNewEncryptedIndexPerm); err != nil {
		return client.EncryptedIndexDescription{}, err
	}

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return client.EncryptedIndexDescription{}, err
	}

	defer txn.Discard()

	index, err := c.newEncryptedIndex(ctx, addRequest)
	if err != nil {
		return client.EncryptedIndexDescription{}, err
	}

	return index, txn.Commit()
}

func (c *collection) newEncryptedIndex(
	ctx context.Context,
	encryptedIndex client.EncryptedIndexDescription,
) (client.EncryptedIndexDescription, error) {
	if encryptedIndex.Type == "" {
		encryptedIndex.Type = client.EncryptedIndexTypeEquality
	}
	err := validateNewEncryptedIndex(c.Version(), encryptedIndex)
	if err != nil {
		return client.EncryptedIndexDescription{}, err
	}

	c.def.EncryptedIndexes = append(c.def.EncryptedIndexes, encryptedIndex)

	err = description.SaveCollection(ctx, c.def)
	if err != nil {
		c.def.EncryptedIndexes = c.def.EncryptedIndexes[:len(c.def.EncryptedIndexes)-1]
		return client.EncryptedIndexDescription{}, err
	}

	err = c.db.loadCollectionDefinitions(ctx)
	if err != nil {
		return client.EncryptedIndexDescription{}, err
	}

	return c.def.EncryptedIndexes[len(c.def.EncryptedIndexes)-1], nil
}

// ListEncryptedIndexes returns all the encrypted indexes that exist on the collection.
func (c *collection) ListEncryptedIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListCollectionEncryptedIndexesOptions],
) ([]client.EncryptedIndexDescription, error) {
	opt := utils.NewOptions(opts...)
	ident := opt.GetIdentity()
	if err := c.db.checkNodeAccess(ctx, ident, acpTypes.NodeListEncryptedIndexPerm); err != nil {
		return nil, err
	}
	return c.Version().EncryptedIndexes, nil
}

// DeleteEncryptedIndex deletes an encrypted index from the collection.
//
// The encrypted index will be removed from the system store.
// All SE artifacts on remote nodes will become inaccessible for queries.
func (c *collection) DeleteEncryptedIndex(
	ctx context.Context,
	fieldName string,
	opts ...options.Enumerable[options.DeleteEncryptedIndexOptions],
) error {
	ctx, _, _ = getTxnAndSetCtxForCollection(ctx, c)

	ctx, span := tracer.Start(ctx)
	defer span.End()

	opt := utils.NewOptions(opts...)
	ident := opt.GetIdentity()

	if err := c.db.checkNodeAccess(ctx, ident, acpTypes.NodeDeleteEncryptedIndexPerm); err != nil {
		return err
	}

	ctx, txn, err := ensureContextTxn(ctx, c.db, false)
	if err != nil {
		return err
	}

	defer txn.Discard()

	err = c.deleteEncryptedIndex(ctx, fieldName)
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (c *collection) deleteEncryptedIndex(ctx context.Context, fieldName string) error {
	indexToRemove := -1
	for i, encIdx := range c.Version().EncryptedIndexes {
		if encIdx.FieldName == fieldName {
			indexToRemove = i
			break
		}
	}

	if indexToRemove == -1 {
		return NewErrEncryptedIndexDoesNotExist(fieldName)
	}

	oldEncryptedIndexes := make([]client.EncryptedIndexDescription, len(c.Version().EncryptedIndexes))
	copy(oldEncryptedIndexes, c.Version().EncryptedIndexes)

	c.def.EncryptedIndexes = append(
		c.def.EncryptedIndexes[:indexToRemove],
		c.def.EncryptedIndexes[indexToRemove+1:]...,
	)

	err := description.SaveCollection(ctx, c.def)
	if err != nil {
		c.def.EncryptedIndexes = oldEncryptedIndexes
		return err
	}

	err = c.db.loadCollectionDefinitions(ctx)
	if err != nil {
		return err
	}

	return nil
}

// checkExistingFieldsAndAdjustRelFieldNames checks if the fields in the index description
// exist in the collection definition.
// If a field is a relation, it will be adjusted to relation id field name, a.k.a. `field_name + _id`.
func checkExistingFieldsAndAdjustRelFieldNames(
	collection client.CollectionVersion,
	fields []client.IndexedFieldDescription,
) error {
	for i := range fields {
		field, found := collection.GetFieldByName(fields[i].Name)
		if !found {
			return NewErrNonExistingFieldForIndex(fields[i].Name)
		}
		if field.Kind.IsObject() {
			fields[i].Name = request.ToFieldID(fields[i].Name)
		}
	}
	return nil
}

// validateNewEncryptedIndex validates, if encrypted index can be added to the given collection.
// It checks if the field exists in the collection definition and if an encrypted index already exists on the field.
func validateNewEncryptedIndex(
	definition client.CollectionVersion,
	newEncryptedIndex client.EncryptedIndexDescription,
) error {
	_, found := definition.GetFieldByName(newEncryptedIndex.FieldName)
	if !found {
		return NewErrEncryptedIndexOnNonExistentField(newEncryptedIndex.FieldName)
	}
	for _, encryptedIndex := range definition.EncryptedIndexes {
		if encryptedIndex.FieldName == newEncryptedIndex.FieldName {
			return NewErrEncryptedIndexAlreadyExists(newEncryptedIndex.FieldName)
		}
	}
	return nil
}

// validateEncryptedIndexesOnCollection validates all encrypted indexes on the collection.
// It checks if the all indexes are set on existing distinct fields.
func validateEncryptedIndexesOnCollection(definition client.CollectionVersion) error {
	encryptedFieldNames := make(map[string]struct{}, len(definition.EncryptedIndexes))
	for _, encryptedIndex := range definition.EncryptedIndexes {
		if _, found := definition.GetFieldByName(encryptedIndex.FieldName); !found {
			return NewErrEncryptedIndexOnNonExistentField(encryptedIndex.FieldName)
		}
		if _, found := encryptedFieldNames[encryptedIndex.FieldName]; found {
			return NewErrEncryptedIndexAlreadyExists(encryptedIndex.FieldName)
		}
		encryptedFieldNames[encryptedIndex.FieldName] = struct{}{}
	}
	return nil
}

func generateIndexNameIfNeeded(
	colVersion client.CollectionVersion,
	newReq client.NewIndexRequest,
) (string, error) {
	indexName := newReq.Name
	if indexName == "" {
		nameIncrement := 1
		for {
			var err error
			indexName, err = generateIndexName(colVersion.Name, newReq.Fields, nameIncrement)
			if err != nil {
				return "", err
			}

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

func validateIndexDescription(desc client.NewIndexRequest) error {
	if desc.Name != "" && !schema.IsValidIndexName(desc.Name) {
		return schema.NewErrIndexWithInvalidName(desc.Name)
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

func generateIndexName(colName string, fields []client.IndexedFieldDescription, inc int) (string, error) {
	sb := strings.Builder{}
	// at the moment we support only single field indexes that can be stored only in
	// ascending order. This will change once we introduce composite indexes.
	direction := "ASC"
	_, err := sb.WriteString(colName)
	if err != nil {
		return "", err
	}

	err = sb.WriteByte('_')
	if err != nil {
		return "", err
	}

	// we can safely assume that there is at least one field in the slice
	// because we validate it before calling this function
	_, err = sb.WriteString(fields[0].Name)
	if err != nil {
		return "", err
	}

	err = sb.WriteByte('_')
	if err != nil {
		return "", err
	}

	_, err = sb.WriteString(direction)
	if err != nil {
		return "", err
	}

	if inc > 1 {
		err = sb.WriteByte('_')
		if err != nil {
			return "", err
		}

		_, err = sb.WriteString(strconv.Itoa(inc))
		if err != nil {
			return "", err
		}
	}

	return sb.String(), nil
}

// listAllEncryptedIndexDescriptions returns all encrypted index descriptions in the database.
func (db *DB) listAllEncryptedIndexDescriptions(
	ctx context.Context,
) (map[client.CollectionName][]client.EncryptedIndexDescription, error) {
	collections, err := description.GetCollections(ctx)

	if err != nil {
		return nil, err
	}

	indexes := make(map[client.CollectionName][]client.EncryptedIndexDescription)

	for _, col := range collections {
		if len(col.EncryptedIndexes) > 0 {
			indexes[col.Name] = col.EncryptedIndexes
		}
	}

	return indexes, nil
}

// reindexNewActiveVersion reindexes all documents in the collection for the new active version.
func (db *DB) reindexNewActiveVersion(ctx context.Context, col client.CollectionVersion) error {
	if !col.IsActive {
		return nil
	}

	txnOpt := datastore.CtxTryGetClientTxnOption(ctx)
	collection, err := db.newCollection(col, txnOpt)
	if err != nil {
		return err
	}
	for _, colIndex := range collection.indexes {
		err = colIndex.RemoveAll(ctx)
		if err != nil {
			return err
		}
		err = collection.indexExistingDocs(ctx, colIndex)
		if err != nil {
			return err
		}
	}

	return nil
}
