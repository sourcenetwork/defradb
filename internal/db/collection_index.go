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

	"github.com/sourcenetwork/corelog"
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

// listIndexDescriptions returns all index descriptions in the database joined with their runtime state.
func (db *DB) listIndexDescriptions(
	ctx context.Context,
) (map[client.CollectionName][]client.ListIndexesResult, error) {
	collections, err := description.GetCollections(ctx, db.collectionRepository)
	if err != nil {
		return nil, err
	}

	indexes := make(map[client.CollectionName][]client.ListIndexesResult)

	for _, col := range collections {
		if len(col.Indexes) == 0 {
			continue
		}
		states, err := getIndexStates(ctx, col.CollectionID)
		if err != nil {
			return nil, err
		}
		results := make([]client.ListIndexesResult, len(col.Indexes))
		for i, desc := range col.Indexes {
			state, ok := states[desc.ID]
			results[i] = state.listResult(col.CollectionID, desc, ok)
		}
		indexes[col.Name] = results
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
// Existing documents are backfilled in batched transactions after the definition
// commits, so no single transaction exceeds the storage engine's size limit.
// With an explicit (caller-provided) transaction the backfill runs inside the
// caller's Commit() and its error cannot be returned — a failure is recorded
// on the index state as status "failed" with a reason.
//
// If the backfill fails, the index definition remains in place with a failed status.
// It is not maintained by subsequent writes. Use DeleteIndex to remove it before recreating.
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

	indexDesc, backfill, err := c.newIndex(ctx, desc)
	if err != nil {
		return client.IndexDescription{}, err
	}

	if txn.explicit {
		collectionID := c.def.CollectionID
		indexID := indexDesc.ID
		txn.OnSuccess(func() {
			if err := backfill(ctx); err != nil {
				log.ErrorE("deferred index backfill failed", err,
					corelog.String("collectionID", collectionID),
					corelog.Any("indexID", indexID),
				)
			}
		})
		return indexDesc, nil
	}

	if err := txn.Commit(); err != nil {
		return client.IndexDescription{}, err
	}

	if err := backfill(ctx); err != nil {
		return client.IndexDescription{}, err
	}
	return indexDesc, nil
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

	if _, err := allocateIndexEpoch(ctx, def.CollectionID, uint32(indexID)); err != nil {
		return client.IndexDescription{}, err
	}

	return client.IndexDescription{
		Name:   indexName,
		ID:     uint32(indexID),
		Fields: desc.Fields,
		Unique: desc.Unique,
	}, nil
}

// allocateIndexEpoch advances the index's epoch sequence and returns the new epoch.
func allocateIndexEpoch(ctx context.Context, collectionID string, indexID uint32) (uint32, error) {
	seq, err := sequence.Get(ctx, keys.NewIndexEpochSequenceKey(collectionID, indexID))
	if err != nil {
		return 0, err
	}
	next, err := seq.Next(ctx)
	if err != nil {
		return 0, err
	}
	return uint32(next), nil
}

// newIndex stages the index definition in the current transaction and returns a
// backfill function the caller must run after the transaction commits.
//
// c.def.Indexes and c.indexes are updated so writes on this collection instance
// maintain the new index immediately; both are rolled back if staging fails.
func (c *collection) newIndex(
	ctx context.Context,
	newReq client.NewIndexRequest,
) (client.IndexDescription, func(context.Context) error, error) {
	desc, err := processNewIndexRequest(ctx, c.Version(), newReq)
	if err != nil {
		return client.IndexDescription{}, nil, err
	}

	c.def.Indexes = append(c.def.Indexes, desc)

	err = description.SaveCollection(ctx, c.db.collectionRepository, c.def)
	if err != nil {
		c.def.Indexes = c.def.Indexes[:len(c.def.Indexes)-1]
		return client.IndexDescription{}, nil, err
	}

	// This registers the build without an "already in progress" guard, which is safe because
	// the index ID keying the record is sequence-allocated and therefore unique per build; a
	// caller that reused an index ID would need to add one.
	err = c.db.startIndexBuild(ctx, c.def.CollectionID, desc.ID)
	if err != nil {
		c.def.Indexes = c.def.Indexes[:len(c.def.Indexes)-1]
		return client.IndexDescription{}, nil, err
	}

	// building=true: this instance maintains the index through the backfill that runs
	// after commit, so writes in that window must use the build-tolerant save/delete.
	colIndex, err := NewCollectionIndex(ctx, c, desc, true)
	if err != nil {
		c.def.Indexes = c.def.Indexes[:len(c.def.Indexes)-1]
		return client.IndexDescription{}, nil, err
	}
	c.indexes = append(c.indexes, colIndex)

	if c.indexStates == nil {
		c.indexStates = make(map[uint32]indexState)
	}
	c.indexStates[desc.ID] = indexState{Action: client.BackfillIndexAction, Status: client.InProgressActionStatus}

	// Backfill builds a fresh collection from this snapshot per batch,
	// so each retry re-reads documents.
	defSnapshot := c.def

	backfill := func(bfCtx context.Context) error {
		return c.db.backfillIndex(bfCtx, defSnapshot, desc, immutable.None[string]())
	}

	return desc, backfill, nil
}

func (c *collection) appendNewIndexAndIndexExistingDocs(
	ctx context.Context,
	desc client.IndexDescription,
) (CollectionIndex, error) {
	colIndex, err := NewCollectionIndex(ctx, c, desc, false)
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

// collectDocIDsAfter performs a keys-only raw range scan over the datastore and collects
// up to limit distinct docIDs in key order: all of them when watermark is None, or only
// those sorting strictly after it.
//
// The scan range covers only active-value keys for the collection, so deleted docs and
// non-document keys are never visited.
func (c *collection) collectDocIDsAfter(
	ctx context.Context,
	shortID uint32,
	watermark immutable.Option[string],
	limit int,
) (docIDs []string, err error) {
	txn := datastore.CtxMustGetTxn(ctx)

	var startKey datastore.Key = keys.DataStoreKey{
		CollectionShortID: shortID,
		InstanceType:      keys.ValueKey,
	}
	if watermark.HasValue() {
		startKey = keys.DataStoreKey{
			CollectionShortID: shortID,
			InstanceType:      keys.ValueKey,
			DocID:             watermark.Value(),
		}.PrefixEnd()
	}

	endKey := keys.DataStoreKey{
		CollectionShortID: shortID,
		InstanceType:      keys.ValueKey,
	}.PrefixEnd()

	iter, err := txn.Datastore().Iterator(ctx, datastore.IterOptions{
		Start:    startKey,
		End:      endKey,
		KeysOnly: true,
	})
	if err != nil {
		return nil, err
	}

	var prevDocID string
	for len(docIDs) < limit {
		hasNext, err := iter.Next()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		dsKey, err := keys.NewDataStoreKey(string(iter.Key()))
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		if dsKey.DocID == "" || dsKey.DocID == prevDocID {
			continue
		}

		prevDocID = dsKey.DocID
		docIDs = append(docIDs, dsKey.DocID)
	}

	return docIDs, iter.Close()
}

// iterateDocsBatch iterates a batch of the collection's documents in docID order.
//
// When limit > 0 a keys-only scan collects the batch's docIDs (after startAfter, if set),
// and the document fetcher is started with one exact per-doc prefix per collected docID.
// Progress (lastDocID, count) is based on those candidate docIDs, not on what the fetcher
// yields, so documents filtered out by ACP cannot stall or truncate the backfill.
//
// When limit == 0 the fetcher scans the whole collection and progress is based on
// fetched documents.
func (c *collection) iterateDocsBatch(
	ctx context.Context,
	fields []client.CollectionFieldDescription,
	startAfter immutable.Option[string],
	limit int,
	exec func(doc *client.Document) error,
) (lastDocID string, count int, err error) {
	shortID, idErr := id.GetShortCollectionID(ctx, c.Version().CollectionID)
	if idErr != nil {
		return "", 0, idErr
	}

	var prefixes []keys.Walkable

	if limit > 0 {
		candidates, scanErr := c.collectDocIDsAfter(ctx, shortID, startAfter, limit)
		if scanErr != nil {
			return "", 0, scanErr
		}
		if len(candidates) == 0 {
			return "", 0, nil
		}

		lastDocID = candidates[len(candidates)-1]
		count = len(candidates)

		prefixes = make([]keys.Walkable, len(candidates))
		for i, docID := range candidates {
			prefixes[i] = keys.DataStoreKey{
				CollectionShortID: shortID,
				DocID:             docID,
			}
		}
	} else {
		prefixes = []keys.Walkable{
			keys.DataStoreKey{CollectionShortID: shortID},
		}
	}

	txn := datastore.CtxMustGetTxn(ctx)
	df := c.newFetcher(ctx)
	initErr := df.Init(
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
	if initErr != nil {
		return "", 0, errors.Join(initErr, df.Close())
	}

	startErr := df.Start(ctx, prefixes...)
	if startErr != nil {
		return "", 0, errors.Join(startErr, df.Close())
	}

	for {
		encodedDoc, _, fetchErr := df.FetchNext(ctx)
		if fetchErr != nil {
			return "", 0, errors.Join(fetchErr, df.Close())
		}
		if encodedDoc == nil {
			break
		}

		doc, decodeErr := fetcher.Decode(ctx, encodedDoc, c.Version())
		if decodeErr != nil {
			return "", 0, errors.Join(decodeErr, df.Close())
		}

		execErr := exec(doc)
		if execErr != nil {
			return "", 0, errors.Join(execErr, df.Close())
		}

		if limit == 0 {
			// Whole-collection mode: track progress from fetched docs.
			lastDocID = string(encodedDoc.ID())
			count++
		}
	}

	return lastDocID, count, df.Close()
}

// iterateAllDocs iterates all documents in the collection in docID order,
// calling exec for each one.  It is a thin wrapper around iterateDocsBatch.
func (c *collection) iterateAllDocs(
	ctx context.Context,
	fields []client.CollectionFieldDescription,
	exec func(doc *client.Document) error,
) error {
	_, _, err := c.iterateDocsBatch(ctx, fields, immutable.None[string](), 0, exec)
	return err
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
// The definition is removed and a dropping state record is written in a single
// transaction. Index entries are then deleted in batched transactions so that no
// single transaction exceeds the storage engine's transaction size limit. With an
// explicit caller-provided transaction the entry GC runs inside the caller's Commit.
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

	gc, err := c.deleteIndex(ctx, indexName)
	if err != nil {
		return err
	}

	return commitAndRunDeferred(ctx, txn, []func(context.Context) error{gc})
}

// deleteIndex stages the index deletion in the current transaction and returns a
// deferred closure that performs the batched GC of index entries. The definition
// is removed and a dropping state record is written in the caller's transaction;
// the returned closure must be run after that transaction commits.
func (c *collection) deleteIndex(ctx context.Context, indexName string) (func(context.Context) error, error) {
	// Locate the description by name in the version definition (source of truth).
	// Failed indexes are excluded from c.indexes but still live in c.Version().Indexes.
	var desc *client.IndexDescription
	for i := range c.Version().Indexes {
		if c.Version().Indexes[i].Name == indexName {
			d := c.Version().Indexes[i]
			desc = &d
			break
		}
	}
	if desc == nil {
		return nil, NewErrIndexWithNameDoesNotExists(indexName)
	}

	// Remove the definition so the planner and writers immediately stop seeing it.
	oldIndexes := make([]client.IndexDescription, len(c.Version().Indexes))
	copy(oldIndexes, c.Version().Indexes)
	for i := range c.Version().Indexes {
		if c.Version().Indexes[i].Name == indexName {
			c.def.Indexes = slices.Delete(c.Version().Indexes, i, i+1)
			break
		}
	}

	if err := description.SaveCollection(ctx, c.db.collectionRepository, c.def); err != nil {
		c.def.Indexes = oldIndexes
		return nil, err
	}

	// Record a dropping state so startup recovery can resume if the process exits
	// before GC completes.
	if err := c.db.startIndexDrop(ctx, c.def.CollectionID, desc.ID); err != nil {
		c.def.Indexes = oldIndexes
		return nil, err
	}

	// Resolve the short collection ID now, while the staging transaction is live,
	// so the deferred GC needs no transaction of its own to look it up.
	shortID, err := id.GetShortCollectionID(ctx, c.def.CollectionID)
	if err != nil {
		c.def.Indexes = oldIndexes
		return nil, err
	}

	for i := range c.indexes {
		if c.indexes[i].Name() == indexName {
			c.indexes = slices.Delete(c.indexes, i, i+1)
			break
		}
	}
	delete(c.indexStates, desc.ID)

	collectionID := c.def.CollectionID
	indexID := desc.ID
	gc := func(gcCtx context.Context) error {
		return c.db.gcIndex(gcCtx, collectionID, shortID, indexID, indexName)
	}

	return gc, nil
}

// ListIndexes returns all indexes for the collection with their current status.
func (c *collection) ListIndexes(
	ctx context.Context,
	opts ...options.Enumerable[options.ListCollectionIndexesOptions],
) ([]client.ListIndexesResult, error) {
	ctx, _, _ = getTxnAndSetCtxForCollection(ctx, c)

	opt := utils.NewOptions(opts...)

	if err := c.db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeListIndexPerm); err != nil {
		return nil, err
	}

	ctx, txn, err := ensureContextTxn(ctx, c.db, true)
	if err != nil {
		return nil, err
	}
	defer txn.Discard()

	states, err := getIndexStates(ctx, c.def.CollectionID)
	if err != nil {
		return nil, err
	}

	indexes := c.Version().Indexes
	result := make([]client.ListIndexesResult, len(indexes))
	for i, desc := range indexes {
		state, ok := states[desc.ID]
		result[i] = state.listResult(c.def.CollectionID, desc, ok)
	}
	return result, nil
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

	err = description.SaveCollection(ctx, c.db.collectionRepository, c.def)
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

	err := description.SaveCollection(ctx, c.db.collectionRepository, c.def)
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
	collections, err := description.GetCollections(ctx, db.collectionRepository)

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

// reindexNewActiveVersion stages a rebuild of every index in the new active version on the
// transaction bound to ctx, so the staging commits with the version switch, and returns a
// function that runs the rebuilds. It must run after the commit, as each rebuild drives its own
// batched transactions. A building index is excluded from query planning, so queries full scan
// until the rebuild completes and flips it back to ready.
func (db *DB) reindexNewActiveVersion(
	ctx context.Context,
	col client.CollectionVersion,
) (func(context.Context) error, error) {
	if !col.IsActive {
		return func(context.Context) error { return nil }, nil
	}

	type rebuild struct {
		desc          client.IndexDescription
		buildingEpoch uint32
		oldEpoch      uint32
	}
	rebuilds := make([]rebuild, 0, len(col.Indexes))

	for _, desc := range col.Indexes {
		// The current sequence value is the epoch being superseded; the next value is the
		// disjoint epoch the rebuild fills.
		oldEpoch, err := getIndexEpoch(ctx, col.CollectionID, desc.ID)
		if err != nil {
			return nil, err
		}

		buildingEpoch, err := allocateIndexEpoch(ctx, col.CollectionID, desc.ID)
		if err != nil {
			return nil, err
		}

		err = db.startIndexRebuild(ctx, col.CollectionID, desc.ID, buildingEpoch, oldEpoch)
		if err != nil {
			return nil, err
		}

		rebuilds = append(rebuilds, rebuild{desc: desc, buildingEpoch: buildingEpoch, oldEpoch: oldEpoch})
	}

	run := func(runCtx context.Context) error {
		for _, r := range rebuilds {
			err := db.runIndexRebuild(runCtx, col, r.desc, immutable.None[string](), r.buildingEpoch, r.oldEpoch)
			if err != nil {
				return err
			}
		}
		return nil
	}
	return run, nil
}
