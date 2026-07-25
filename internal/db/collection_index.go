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
	"encoding/binary"
	"maps"
	"strconv"
	"strings"

	"github.com/sourcenetwork/corekv"
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
		states, err := getIndexBuildStates(ctx, col.CollectionID)
		if err != nil {
			return nil, err
		}
		results := make([]client.ListIndexesResult, len(col.Indexes))
		for i, desc := range col.Indexes {
			state, ok := states[desc.ID]
			results[i] = state.listResult(col.CollectionID, col.Name, desc, ok)
		}
		indexes[col.Name] = results
	}

	return indexes, nil
}

// currentIndexes resolves the indexes a write must maintain, on the write's own transaction.
//
// A handle's cached c.indexes is a snapshot from construction and goes stale when an index is
// created, dropped, or rebuilt. Each of those bumps the collection's index-mutation counter (create
// and rebuild via the epoch advance, drop via the drop staging). A write reads the counter: if it
// matches the value the cache was resolved at, the cache is returned as is; otherwise it re-resolves
// and caches the fresh instances and counter.
//
// The read also keeps it correct under concurrency: the counter key is in the write's read-set, so a
// concurrent index change that bumped it conflicts the commit, and withTxnRetries re-runs against the
// new counter.
//
// A build finishing (building->ready) deliberately does not bump. The building flag only widens the
// unique-index same-doc and delete-of-missing tolerances, so a stale-building handle still rejects a
// real cross-doc duplicate; not bumping also keeps the worker's completion from conflicting live index
// creates.
func (c *collection) currentIndexes(ctx context.Context) ([]client.CollectionIndex, error) {
	collectionShortID, err := c.collectionShortID(ctx)
	if err != nil {
		return nil, err
	}

	counter, err := readIndexMutationCounter(ctx, collectionShortID)
	if err != nil {
		return nil, err
	}

	if counter == c.indexesCounter {
		return c.indexes, nil
	}

	indexes, err := c.resolveIndexes(ctx)
	if err != nil {
		return nil, err
	}

	c.indexes = indexes
	c.indexesCounter = counter
	return indexes, nil
}

// collectionShortID returns the collection's short ID, resolving it once and caching it. The short ID
// is stable for the collection's life, so the write path reads it from storage only on the first call.
func (c *collection) collectionShortID(ctx context.Context) (uint32, error) {
	if c.shortID != 0 {
		return c.shortID, nil
	}
	shortID, err := id.GetCollectionShortID(ctx, c.def.CollectionID)
	if err != nil {
		return 0, err
	}
	c.shortID = shortID
	return shortID, nil
}

// resolveIndexes builds the write-path index instances from the current definition and live epochs.
//
// The per-index building/failed status lives in action-status records. In the steady state no index
// has a record (a completed build deletes its own), which means every index is ready: a keys-only
// probe that stops at the first record confirms this without decoding anything, and the instances
// are built ready (building=false). Only when the probe finds a record (a build or drop is in
// flight) is the full state decode run to get each index's real building/failed status. The epoch is
// always read live, since a rebuild advances it while the index id is unchanged.
func (c *collection) resolveIndexes(ctx context.Context) ([]client.CollectionIndex, error) {
	def, err := description.GetCollectionByID(ctx, c.db.collectionRepository, c.def.VersionID)
	if err != nil {
		return nil, err
	}

	if len(def.Indexes) == 0 {
		return nil, nil
	}

	hasActions, err := hasAnyIndexAction(ctx, def.CollectionID)
	if err != nil {
		return nil, err
	}

	// Only decode the per-index status when a record exists. With none, every index is ready:
	// building=false is correct (a ready index carries no record) and none is failed.
	var states map[uint32]indexState
	if hasActions {
		states, err = getIndexBuildStates(ctx, def.CollectionID)
		if err != nil {
			return nil, err
		}
	}

	indexes := make([]client.CollectionIndex, 0, len(def.Indexes))
	for _, index := range def.Indexes {
		state := states[index.ID]
		// A failed index is abandoned and not maintained by writes. A building index is maintained
		// so its epoch keeps receiving concurrent writes while the backfill runs. The building flag
		// only affects correctness for a unique index mid-build (saveUniqueKey tolerates a concurrent
		// entry when building), so getting the real status when a record exists is what matters.
		if state.isFailed() {
			continue
		}
		colIndex, err := c.db.newLoadedCollectionIndex(ctx, c, index, state.isBuilding())
		if err != nil {
			return nil, err
		}
		// A kind this build does not implement was recorded as failed and skipped.
		if colIndex == nil {
			continue
		}
		indexes = append(indexes, colIndex)
	}
	return indexes, nil
}

// readIndexMutationCounter returns the collection's current index-mutation counter, or 0 if it has
// never been bumped. It is a plain read (not sequence.Get, which seeds the key on a miss) so the write
// hot path never writes; the read still puts the key in the caller's read-set, which is what makes a
// concurrent bump conflict the write and retry.
func readIndexMutationCounter(ctx context.Context, collectionShortID uint32) (uint64, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	val, err := txn.Systemstore().Get(ctx, keys.NewIndexMutationCounterKey(collectionShortID).Bytes())
	if errors.Is(err, corekv.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(val), nil
}

// bumpIndexMutationCounter advances the collection's index-mutation counter on the transaction bound
// to ctx. Called wherever the write-maintained index set changes: index create and rebuild (the epoch
// advance) and drop. It advances a sequence, which reads the counter and writes it back, so a
// concurrent bump or an in-flight write that read the old value conflicts, keeping writes maintaining
// the right index set across a concurrent change.
func bumpIndexMutationCounter(ctx context.Context, collectionShortID uint32) error {
	seq, err := sequence.Get(ctx, keys.NewIndexMutationCounterKey(collectionShortID))
	if err != nil {
		return err
	}
	_, err = seq.Next(ctx)
	return err
}

// hasAnyIndexAction reports whether the collection has any index action-status record (a build or
// drop in flight). It is a keys-only probe that stops at the first key, so it decodes nothing; the
// steady state (no records = every index ready) pays only this probe, not the full state decode.
func hasAnyIndexAction(ctx context.Context, collectionID string) (bool, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	iter, err := txn.Systemstore().Iterator(ctx, corekv.IterOptions{
		Prefix:   indexActionCollectionPrefix(collectionID),
		KeysOnly: true,
	})
	if err != nil {
		return false, err
	}
	hasValue, err := iter.Next()
	if err != nil {
		return false, errors.Join(err, iter.Close())
	}
	return hasValue, iter.Close()
}

func (c *collection) updateDocIndex(ctx context.Context, oldDoc, newDoc *client.Document) error {
	indexes, err := c.currentIndexes(ctx)
	if err != nil {
		return err
	}
	if err := c.deleteFromIndexes(ctx, indexes, oldDoc); err != nil {
		return err
	}
	return c.saveToIndexes(ctx, indexes, newDoc)
}

func (c *collection) addDocToIndex(ctx context.Context, doc *client.Document) error {
	// callers of this function must set a context transaction
	indexes, err := c.currentIndexes(ctx)
	if err != nil {
		return err
	}
	return c.saveToIndexes(ctx, indexes, doc)
}

func (c *collection) saveToIndexes(
	ctx context.Context,
	indexes []client.CollectionIndex,
	doc *client.Document,
) error {
	for _, index := range indexes {
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
	indexes, err := c.currentIndexes(ctx)
	if err != nil {
		return err
	}
	if len(indexes) == 0 {
		return nil
	}

	primaryKey, err := c.getPrimaryKeyFromDocID(ctx, doc.ID())
	if err != nil {
		return err
	}

	// Collect the fields to fetch from the resolved indexes, not the handle's snapshot. A stale
	// handle does not list a concurrently-created index, so its snapshot would omit that index's
	// field and the old value needed to delete its stale entry would never be fetched.
	oldDoc, err := c.get(
		ctx,
		primaryKey,
		c.collectIndexedFields(indexes),
		false,
	)
	if err != nil {
		return err
	}
	for _, index := range indexes {
		err = index.Update(ctx, oldDoc, doc)
		if err != nil {
			return err
		}
	}
	return nil
}

// collectIndexedFields returns the distinct fields covered by the given indexes, for fetching the
// prior document state a secondary-index update needs. Field lookup uses the collection schema,
// which an index creation does not change, so the handle's definition is fine here.
func (c *collection) collectIndexedFields(
	indexes []client.CollectionIndex,
) []client.CollectionFieldDescription {
	seen := make(map[string]bool)
	fields := make([]client.CollectionFieldDescription, 0, len(indexes))
	for _, index := range indexes {
		for _, field := range index.Description().Fields {
			if seen[field.Name] {
				continue
			}
			seen[field.Name] = true
			colField, ok := c.Version().GetFieldByName(field.Name)
			if ok {
				fields = append(fields, colField)
			}
		}
	}
	return fields
}

func (c *collection) deleteIndexedDoc(
	ctx context.Context,
	doc *client.Document,
) error {
	indexes, err := c.currentIndexes(ctx)
	if err != nil {
		return err
	}
	return c.deleteFromIndexes(ctx, indexes, doc)
}

func (c *collection) deleteFromIndexes(
	ctx context.Context,
	indexes []client.CollectionIndex,
	doc *client.Document,
) error {
	for _, index := range indexes {
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
	indexes, err := c.currentIndexes(ctx)
	if err != nil {
		return err
	}
	if len(indexes) == 0 {
		return nil
	}

	primaryKey, err := c.getPrimaryKeyFromDocID(ctx, docID)
	if err != nil {
		return err
	}

	// we need to fetch the document to delete it from the indexes, because in order to do so
	// we need to know the values of the fields that are indexed. Fields come from the resolved
	// indexes so a stale handle still fetches a concurrently-created index's field.
	doc, err := c.get(
		ctx,
		primaryKey,
		c.collectIndexedFields(indexes),
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
	return c.deleteFromIndexes(ctx, indexes, doc)
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
// See the client.Collection interface for the async build contract. Mechanically: this commits the
// definition and a "building" state record, which wakes the background worker to run the backfill;
// it does not wait. Writes on this instance maintain the building index, so they and the backfill
// converge on the same final entries.
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

	// With an explicit transaction the caller owns the commit (and any retry). Stage on it and return;
	// the caller's Commit() commits the build record and wakes the worker.
	if txn.explicit {
		defer txn.Discard()
		return c.newIndex(ctx, desc)
	}
	txn.Discard()

	// Implicit transaction: stage and commit under retry. Allocating the epoch bumps the collection's
	// index-mutation counter, which another concurrent index change also writes, so the commit can
	// legitimately conflict; retry re-stages from a clean slate. The staged in-memory state on the
	// handle is restored before each attempt so a retry does not stack a second index.
	baseIndexes := slices.Clone(c.def.Indexes)
	baseColIndexes := slices.Clone(c.indexes)
	baseStates := maps.Clone(c.indexBuildStates)

	var indexDesc client.IndexDescription
	err = c.db.withTxnRetries(ctx, func(attemptCtx context.Context) error {
		c.def.Indexes = slices.Clone(baseIndexes)
		c.indexes = slices.Clone(baseColIndexes)
		c.indexBuildStates = maps.Clone(baseStates)
		var err error
		indexDesc, err = c.newIndex(attemptCtx, desc)
		return err
	})
	if err != nil {
		c.def.Indexes = baseIndexes
		c.indexes = baseColIndexes
		c.indexBuildStates = baseStates
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
		Name:    indexName,
		ID:      uint32(indexID),
		Fields:  desc.Fields,
		Unique:  desc.Unique,
		Kind:    desc.Kind,
		Options: desc.Options,
	}, nil
}

// allocateIndexEpoch advances the index's epoch sequence and returns the new epoch.
func allocateIndexEpoch(ctx context.Context, collectionID string, indexID uint32) (uint32, error) {
	collectionShortID, err := id.GetCollectionShortID(ctx, collectionID)
	if err != nil {
		return 0, err
	}
	seq, err := sequence.Get(ctx, keys.NewIndexEpochSequenceKey(collectionShortID, indexID))
	if err != nil {
		return 0, err
	}
	next, err := seq.Next(ctx)
	if err != nil {
		return 0, err
	}

	// Advancing the epoch changes what a write must maintain (create allocates a first epoch, rebuild
	// advances it), so bump the counter here to cover both in one place.
	if err := bumpIndexMutationCounter(ctx, collectionShortID); err != nil {
		return 0, err
	}
	return uint32(next), nil
}

// newIndex stages the definition and a "building" state record in the current transaction. The
// commit publishes a build event that wakes the worker, which backfills existing documents.
//
// c.def.Indexes and c.indexes are updated so writes on this instance maintain the new index at
// once; both are rolled back if staging fails.
func (c *collection) newIndex(
	ctx context.Context,
	newReq client.NewIndexRequest,
) (client.IndexDescription, error) {
	desc, err := processNewIndexRequest(ctx, c.Version(), newReq)
	if err != nil {
		return client.IndexDescription{}, err
	}

	c.def.Indexes = append(c.def.Indexes, desc)

	err = description.SaveCollection(ctx, c.db.collectionRepository, c.def)
	if err != nil {
		c.def.Indexes = c.def.Indexes[:len(c.def.Indexes)-1]
		return client.IndexDescription{}, err
	}

	// This registers the build without an "already in progress" guard, which is safe because
	// the index ID keying the record is sequence-allocated and therefore unique per build; a
	// caller that reused an index ID would need to add one.
	err = c.db.startIndexBuild(ctx, c.def.CollectionID, desc.ID)
	if err != nil {
		c.def.Indexes = c.def.Indexes[:len(c.def.Indexes)-1]
		return client.IndexDescription{}, err
	}

	// building=true: this instance maintains the index through the background backfill, so
	// writes in that window must use the build-tolerant save/delete.
	colIndex, err := NewCollectionIndex(ctx, c, desc, true)
	if err != nil {
		c.def.Indexes = c.def.Indexes[:len(c.def.Indexes)-1]
		return client.IndexDescription{}, err
	}
	c.indexes = append(c.indexes, colIndex)

	if c.indexBuildStates == nil {
		c.indexBuildStates = make(map[uint32]indexState)
	}
	c.indexBuildStates[desc.ID] = indexState{Action: client.BackfillIndexAction, Status: client.InProgressActionStatus}

	return desc, nil
}

func (c *collection) appendNewIndexAndIndexExistingDocs(
	ctx context.Context,
	desc client.IndexDescription,
) (client.CollectionIndex, error) {
	colIndex, err := NewCollectionIndex(ctx, c, desc, false)
	if err != nil {
		return nil, err
	}

	c.indexes = append(c.indexes, colIndex)

	// This runs inside the collection-definition transaction, so a failure here discards every
	// index write along with the rest of that transaction; no explicit rollback is needed.
	if err := c.indexExistingDocs(ctx, colIndex); err != nil {
		return nil, err
	}

	return colIndex, nil
}

// collectDocShortIDsAfter performs a keys-only raw range scan over the datastore and collects
// up to limit distinct document short IDs in key order: all of them when watermark is None, or only
// those sorting strictly after it.
//
// The scan range covers only active-value keys for the collection, so deleted docs and
// non-document keys are never visited.
func (c *collection) collectDocShortIDsAfter(
	ctx context.Context,
	collectionShortID uint32,
	watermark immutable.Option[uint64],
	limit int,
) (docShortIDs []uint64, err error) {
	txn := datastore.CtxMustGetTxn(ctx)

	var startKey datastore.Key = keys.DataStoreKey{
		CollectionShortID: collectionShortID,
		InstanceType:      keys.ValueKey,
	}
	if watermark.HasValue() {
		startKey = keys.DataStoreKey{
			CollectionShortID: collectionShortID,
			InstanceType:      keys.ValueKey,
			DocShortID:        watermark.Value(),
		}.PrefixEnd()
	}

	endKey := keys.DataStoreKey{
		CollectionShortID: collectionShortID,
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

	var prevDocShortID uint64
	for len(docShortIDs) < limit {
		hasNext, err := iter.Next()
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		dsKey, err := keys.DecodeDataStoreKey(iter.Key())
		if err != nil {
			return nil, errors.Join(err, iter.Close())
		}

		if dsKey.DocShortID == 0 || dsKey.DocShortID == prevDocShortID {
			continue
		}

		prevDocShortID = dsKey.DocShortID
		docShortIDs = append(docShortIDs, dsKey.DocShortID)
	}

	return docShortIDs, iter.Close()
}

// iterateDocsBatch iterates a batch of the collection's documents in storage order.
//
// When limit > 0 a keys-only scan collects the batch's document short IDs (after startAfter,
// if set), and the document fetcher is started with one exact per-doc prefix per collected
// document short ID. Progress is based on those candidates, not on what the fetcher yields,
// so documents filtered out by ACP cannot stall or truncate the backfill.
//
// When limit == 0 the fetcher scans the whole collection and progress is based on
// fetched documents.
func (c *collection) iterateDocsBatch(
	ctx context.Context,
	fields []client.CollectionFieldDescription,
	startAfter immutable.Option[uint64],
	limit int,
	exec func(doc *client.Document) error,
) (lastDocShortID uint64, count int, err error) {
	collectionShortID, idErr := c.collectionShortID(ctx)
	if idErr != nil {
		return 0, 0, idErr
	}

	var prefixes []keys.Walkable

	if limit > 0 {
		candidates, scanErr := c.collectDocShortIDsAfter(ctx, collectionShortID, startAfter, limit)
		if scanErr != nil {
			return 0, 0, scanErr
		}
		if len(candidates) == 0 {
			return 0, 0, nil
		}

		lastDocShortID = candidates[len(candidates)-1]
		count = len(candidates)

		prefixes = make([]keys.Walkable, len(candidates))
		for i, docShortID := range candidates {
			prefixes[i] = keys.DataStoreKey{
				CollectionShortID: collectionShortID,
				DocShortID:        docShortID,
			}
		}
	} else {
		prefixes = []keys.Walkable{
			keys.DataStoreKey{CollectionShortID: collectionShortID},
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
		return 0, 0, errors.Join(initErr, df.Close())
	}

	startErr := df.Start(ctx, prefixes...)
	if startErr != nil {
		return 0, 0, errors.Join(startErr, df.Close())
	}

	for {
		encodedDoc, _, fetchErr := df.FetchNext(ctx)
		if fetchErr != nil {
			return 0, 0, errors.Join(fetchErr, df.Close())
		}
		if encodedDoc == nil {
			break
		}

		doc, decodeErr := fetcher.Decode(ctx, encodedDoc, c.Version())
		if decodeErr != nil {
			return 0, 0, errors.Join(decodeErr, df.Close())
		}

		execErr := exec(doc)
		if execErr != nil {
			return 0, 0, errors.Join(execErr, df.Close())
		}

		if limit == 0 {
			count++
		}
	}

	return lastDocShortID, count, df.Close()
}

// iterateAllDocs iterates all documents in the collection in storage order,
// calling exec for each one.  It is a thin wrapper around iterateDocsBatch.
func (c *collection) iterateAllDocs(
	ctx context.Context,
	fields []client.CollectionFieldDescription,
	exec func(doc *client.Document) error,
) error {
	_, _, err := c.iterateDocsBatch(ctx, fields, immutable.None[uint64](), 0, exec)
	return err
}

func (c *collection) indexExistingDocs(
	ctx context.Context,
	index client.CollectionIndex,
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

	// With an explicit transaction the caller owns the commit (and any retry). Stage on it and return;
	// the caller's Commit() commits the drop record and wakes the worker.
	if txn.explicit {
		defer txn.Discard()
		return c.deleteIndex(ctx, indexName)
	}
	txn.Discard()

	// Implicit transaction: stage and commit under retry. The drop bumps the collection's
	// index-mutation counter, which another concurrent index change also writes, so the commit can
	// legitimately conflict; retry re-stages from the restored index set.
	baseIndexes := slices.Clone(c.def.Indexes)
	baseColIndexes := slices.Clone(c.indexes)
	baseStates := maps.Clone(c.indexBuildStates)

	err = c.db.withTxnRetries(ctx, func(attemptCtx context.Context) error {
		c.def.Indexes = slices.Clone(baseIndexes)
		c.indexes = slices.Clone(baseColIndexes)
		c.indexBuildStates = maps.Clone(baseStates)
		return c.deleteIndex(attemptCtx, indexName)
	})
	if err != nil {
		c.def.Indexes = baseIndexes
		c.indexes = baseColIndexes
		c.indexBuildStates = baseStates
		return err
	}
	return nil
}

// deleteIndex stages the deletion in the current transaction: it removes the definition and writes
// a "dropping" state record. The commit publishes a drop event that wakes the worker, which runs
// the batched entry GC and clears the record. The record also lets startup recovery resume the GC
// if the process exits first.
func (c *collection) deleteIndex(ctx context.Context, indexName string) error {
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
		return NewErrIndexWithNameDoesNotExists(indexName)
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
		return err
	}

	// Record a dropping state so the background worker (and startup recovery) runs the GC.
	if err := c.db.startIndexDrop(ctx, c.def.CollectionID, desc.ID); err != nil {
		c.def.Indexes = oldIndexes
		return err
	}

	// Clear any leftover backfill record (building or failed); the drop record alone does not, so it
	// would otherwise orphan once the definition is gone.
	if err := c.db.clearIndexBuildRecord(ctx, c.def.CollectionID, desc.ID); err != nil {
		c.def.Indexes = oldIndexes
		return err
	}

	// A drop removes an index from the set a write must maintain without touching any epoch, so bump
	// the counter here on the same transaction as the definition change.
	collectionShortID, err := c.collectionShortID(ctx)
	if err != nil {
		c.def.Indexes = oldIndexes
		return err
	}
	if err := bumpIndexMutationCounter(ctx, collectionShortID); err != nil {
		c.def.Indexes = oldIndexes
		return err
	}

	for i := range c.indexes {
		if c.indexes[i].Name() == indexName {
			c.indexes = slices.Delete(c.indexes, i, i+1)
			break
		}
	}
	delete(c.indexBuildStates, desc.ID)

	return nil
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

	states, err := getIndexBuildStates(ctx, c.def.CollectionID)
	if err != nil {
		return nil, err
	}

	version := c.Version()
	indexes := version.Indexes
	result := make([]client.ListIndexesResult, len(indexes))
	for i, desc := range indexes {
		state, ok := states[desc.ID]
		result[i] = state.listResult(c.def.CollectionID, version.Name, desc, ok)
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
	return validateIndexKind(desc)
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

// reindexNewActiveVersion stages, for every index in the new active version, a fresh-epoch build on
// the transaction bound to ctx so it commits with the version switch. It advances the epoch sequence,
// marks the old epoch's entries stale, and records a building state; the commit wakes the worker,
// which fills the new epoch and (via the stale marker) collects the old one, all in the background.
// A building index is excluded from query planning, so queries full scan until its build finishes.
func (db *DB) reindexNewActiveVersion(ctx context.Context, col client.CollectionVersion) error {
	if !col.IsActive {
		return nil
	}

	collectionShortID, err := id.GetCollectionShortID(ctx, col.CollectionID)
	if err != nil {
		return err
	}

	for _, desc := range col.Indexes {
		// Advance the sequence to the new epoch the build will fill; reads resolve the epoch
		// from the sequence, so the build needs no epoch in its own record.
		if _, err := allocateIndexEpoch(ctx, col.CollectionID, desc.ID); err != nil {
			return err
		}

		// Advancing the epoch leaves the old one's entries stale. Mark the index so the worker's
		// sweep collects them; the sweep clears the marker once the GC is done.
		if err := db.markStaleEpochs(ctx, collectionShortID, desc.ID); err != nil {
			return err
		}

		if err := db.startIndexBuild(ctx, col.CollectionID, desc.ID); err != nil {
			return err
		}
	}

	return nil
}
