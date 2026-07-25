// Copyright 2026 Democratized Data Foundation
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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// newBM25NameIndex creates a BM25 index on "name" and drains the build, mirroring newNameIndex
// for the ordered key index.
func newBM25NameIndex(
	t *testing.T,
	ctx context.Context,
	db *DB,
	col client.Collection,
	options map[string]any,
) (client.IndexDescription, error) {
	t.Helper()
	desc, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields:  []client.IndexedFieldDescription{{Name: "name"}},
		Kind:    client.IndexKindBM25,
		Options: options,
	})
	if err != nil {
		return desc, err
	}
	db.indexBuildWorker.drainSync(ctx)
	return desc, nil
}

// bm25State is everything one BM25 index holds, read back from storage.
type bm25State struct {
	// documents and totalLength are the collection totals record.
	documents   uint64
	totalLength uint64
	// postings is the number of (term, document) entries, lengths the number of per-document
	// field length entries.
	postings int
	lengths  int
}

// withBM25Index runs f with the index's write-path instance and a read-only transaction, so a
// test can read back the exact keys the index writes.
func withBM25Index(
	t *testing.T,
	ctx context.Context,
	db *DB,
	col client.Collection,
	desc client.IndexDescription,
	f func(index *collectionBM25Index, txnCtx context.Context, collectionShortID uint32),
) {
	t.Helper()
	rawTxn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer rawTxn.Discard()
	txnCtx := InitContext(ctx, rawTxn)

	colIndex, err := NewCollectionIndex(txnCtx, col, desc, false)
	require.NoError(t, err)
	collectionShortID, err := id.GetCollectionShortID(txnCtx, col.Version().CollectionID)
	require.NoError(t, err)

	bm25Index, ok := colIndex.(*collectionBM25Index)
	require.True(t, ok)

	f(bm25Index, txnCtx, collectionShortID)
}

func readBM25State(
	t *testing.T,
	ctx context.Context,
	db *DB,
	col client.Collection,
	desc client.IndexDescription,
) bm25State {
	t.Helper()
	var state bm25State
	withBM25Index(t, ctx, db, col, desc,
		func(index *collectionBM25Index, txnCtx context.Context, collectionShortID uint32) {
			totalsKey := index.key(collectionShortID, core.BM25TotalsPart, "", 0)
			totals, found, err := index.readUvarints(txnCtx, &totalsKey, 2)
			require.NoError(t, err)
			if found {
				state.documents, state.totalLength = totals[0], totals[1]
			}
			state.postings = countKeysUnder(t, txnCtx, index.key(collectionShortID, core.BM25PostingPart, "", 0))
			state.lengths = countKeysUnder(t, txnCtx, index.key(collectionShortID, core.BM25LengthPart, "", 0))
		})
	return state
}

// bm25TermFrequency returns how often a term is recorded as occurring in a document, and whether
// the posting exists at all.
func bm25TermFrequency(
	t *testing.T,
	ctx context.Context,
	db *DB,
	col client.Collection,
	desc client.IndexDescription,
	term string,
	doc *client.Document,
) (uint64, bool) {
	t.Helper()
	var frequency uint64
	var found bool
	withBM25Index(t, ctx, db, col, desc,
		func(index *collectionBM25Index, txnCtx context.Context, collectionShortID uint32) {
			_, docShortID, err := index.docShortIDs(txnCtx, doc)
			require.NoError(t, err)
			key := index.key(collectionShortID, core.BM25PostingPart, term, docShortID)
			values, exists, err := index.readUvarints(txnCtx, &key, 1)
			require.NoError(t, err)
			found = exists
			if exists {
				frequency = values[0]
			}
		})
	return frequency, found
}

// countKeysUnder returns the number of keys stored under a prefix, using the transaction on ctx.
func countKeysUnder(t *testing.T, ctx context.Context, prefix keys.IndexDataStoreKey) int {
	t.Helper()
	txn := datastore.CtxMustGetTxn(ctx)
	iter, err := txn.Datastore().Iterator(ctx, datastore.IterOptions{Prefix: &prefix, KeysOnly: true})
	require.NoError(t, err)

	var count int
	for {
		hasNext, err := iter.Next()
		if err != nil {
			require.NoError(t, errors.Join(err, iter.Close()))
		}
		if !hasNext {
			break
		}
		count++
	}
	require.NoError(t, iter.Close())
	return count
}

// A backfill must write one posting per distinct term of each document, one field length per
// document, and totals covering all of them.
func TestBM25Index_Backfill_WritesPostingsLengthsAndTotals(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	// Six terms, five of them distinct, "the" occurring twice.
	doc := addUserDoc(t, ctx, col, "the cat sat on the mat")
	// Two terms, both distinct.
	addUserDoc(t, ctx, col, "the dog")

	desc, err := newBM25NameIndex(t, ctx, db, col, nil)
	require.NoError(t, err)

	assert.Equal(t, bm25State{documents: 2, totalLength: 8, postings: 7, lengths: 2},
		readBM25State(t, ctx, db, col, desc))

	frequency, found := bm25TermFrequency(t, ctx, db, col, desc, "the", doc)
	require.True(t, found)
	assert.Equal(t, uint64(2), frequency)
}

// A value with no terms at all gets no entries and is not counted, the same way a value with no
// trigrams gets no entries in a trigram index.
func TestBM25Index_Backfill_WithValueThatHasNoTerms_WritesNothing(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	// Punctuation splits into nothing, and a single character is dropped as a term.
	addUserDoc(t, ctx, col, "--- a")

	desc, err := newBM25NameIndex(t, ctx, db, col, nil)
	require.NoError(t, err)

	assert.Equal(t, bm25State{}, readBM25State(t, ctx, db, col, desc))
}

// A live write into an existing index must record what the backfill would have, and an update
// must take the old value's terms and length back out before adding the new ones.
func TestBM25Index_LiveWriteAndUpdate_MaintainStatistics(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	desc, err := newBM25NameIndex(t, ctx, db, col, nil)
	require.NoError(t, err)
	require.Equal(t, bm25State{}, readBM25State(t, ctx, db, col, desc))

	doc := addUserDoc(t, ctx, col, "the cat sat on the mat")
	assert.Equal(t, bm25State{documents: 1, totalLength: 6, postings: 5, lengths: 1},
		readBM25State(t, ctx, db, col, desc))

	// "a" is a single character and is dropped, leaving one term.
	require.NoError(t, doc.Set(ctx, "name", "a dog"))
	require.NoError(t, col.UpdateDocument(ctx, doc))

	assert.Equal(t, bm25State{documents: 1, totalLength: 1, postings: 1, lengths: 1},
		readBM25State(t, ctx, db, col, desc))
	_, found := bm25TermFrequency(t, ctx, db, col, desc, "the", doc)
	assert.False(t, found, "the old value's terms must not survive the update")
	frequency, found := bm25TermFrequency(t, ctx, db, col, desc, "dog", doc)
	require.True(t, found)
	assert.Equal(t, uint64(1), frequency)

	// Updating to a value with no terms leaves the document out of the index entirely.
	require.NoError(t, doc.Set(ctx, "name", ""))
	require.NoError(t, col.UpdateDocument(ctx, doc))
	assert.Equal(t, bm25State{}, readBM25State(t, ctx, db, col, desc))
}

// Deleting a document must leave no posting, no field length and no share of the totals behind.
func TestBM25Index_OnDocumentDelete_RemovesPostingsAndCounts(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	desc, err := newBM25NameIndex(t, ctx, db, col, nil)
	require.NoError(t, err)

	doc := addUserDoc(t, ctx, col, "the cat sat on the mat")
	addUserDoc(t, ctx, col, "the dog")
	require.Equal(t, bm25State{documents: 2, totalLength: 8, postings: 7, lengths: 2},
		readBM25State(t, ctx, db, col, desc))

	deleted, err := col.DeleteDocument(ctx, doc.ID())
	require.NoError(t, err)
	require.True(t, deleted)

	assert.Equal(t, bm25State{documents: 1, totalLength: 2, postings: 2, lengths: 1},
		readBM25State(t, ctx, db, col, desc))
	_, found := bm25TermFrequency(t, ctx, db, col, desc, "cat", doc)
	assert.False(t, found, "the deleted document must leave no postings behind")
}

func TestBM25Index_WithUnique_ReturnsError(t *testing.T) {
	ctx := context.Background()
	_, col := setupUserCollection(t, ctx)

	_, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
		Kind:   client.IndexKindBM25,
		Unique: true,
	})
	require.ErrorIs(t, err, ErrBM25IndexUnique)
}

func TestBM25Index_WithMoreThanOneField_ReturnsError(t *testing.T) {
	ctx := context.Background()
	db := newBadgerDBNoIndexWorker(t, ctx)
	_, err := db.AddCollection(ctx, `type User { name: String, alias: String }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	_, err = col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}, {Name: "alias"}},
		Kind:   client.IndexKindBM25,
	})
	require.ErrorIs(t, err, ErrBM25IndexNotSingleField)
}

func TestBM25Index_OnNonStringField_ReturnsError(t *testing.T) {
	ctx := context.Background()
	db := newBadgerDBNoIndexWorker(t, ctx)
	_, err := db.AddCollection(ctx, `type User { name: String, age: Int }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	_, err = col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "age"}},
		Kind:   client.IndexKindBM25,
	})
	require.ErrorIs(t, err, NewErrBM25IndexFieldNotString("age", "Int"))
	require.ErrorContains(t, err, "age")
}

func TestBM25Index_WithUnknownOption_ReturnsError(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	_, err := newBM25NameIndex(t, ctx, db, col, map[string]any{"k2": 1.2})
	require.ErrorIs(t, err, NewErrBM25IndexUnknownOption("k2"))
}

func TestBM25Index_WithInvalidOptionValue_ReturnsError(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	_, err := newBM25NameIndex(t, ctx, db, col, map[string]any{"k1": "fast"})
	require.ErrorIs(t, err, NewErrBM25IndexInvalidOption("k1", "fast"))

	_, err = newBM25NameIndex(t, ctx, db, col, map[string]any{"b": 2.0})
	require.ErrorIs(t, err, NewErrBM25IndexInvalidOption("b", 2.0))
}
