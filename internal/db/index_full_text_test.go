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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/encoding"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func newBM25NameIndex(
	t *testing.T,
	ctx context.Context,
	db *DB,
	col client.Collection,
) client.IndexDescription {
	t.Helper()
	desc, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
		FullText: &client.FullTextIndexDescription{
			Algorithm: client.FullTextAlgorithmBM25,
			BM25: &client.BM25Params{
				K1: client.DefaultBM25K1,
				B:  client.DefaultBM25B,
			},
		},
	})
	require.NoError(t, err)
	db.indexBuildWorker.drainSync(ctx)
	return desc
}

type fullTextState struct {
	documents   uint64
	totalLength uint64
	postings    int
	lengths     int
}

func readFullTextState(
	t *testing.T,
	ctx context.Context,
	db *DB,
	col client.Collection,
	desc client.IndexDescription,
) fullTextState {
	t.Helper()
	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)
	collectionShortID, err := id.GetCollectionShortID(txnCtx, col.Version().CollectionID)
	require.NoError(t, err)
	epoch, err := getIndexEpoch(txnCtx, col.Version().CollectionID, desc.ID)
	require.NoError(t, err)

	state := fullTextState{}
	totalsKey := keys.NewFullTextTotalsKey(collectionShortID, desc.ID, epoch)
	if data, err := datastore.CtxMustGetTxn(txnCtx).Datastore().Get(txnCtx, &totalsKey); err == nil {
		data, state.documents, err = encoding.DecodeUvarintAscending(data)
		require.NoError(t, err)
		data, state.totalLength, err = encoding.DecodeUvarintAscending(data)
		require.NoError(t, err)
		require.Empty(t, data)
	}
	postingPrefix := keys.NewFullTextPostingKey(collectionShortID, desc.ID, epoch, "", 0)
	lengthPrefix := keys.NewFullTextLengthKey(collectionShortID, desc.ID, epoch, 0)
	state.postings = countFullTextKeys(t, txnCtx, &postingPrefix)
	state.lengths = countFullTextKeys(t, txnCtx, &lengthPrefix)
	return state
}

func countFullTextKeys(t *testing.T, ctx context.Context, prefix keys.Walkable) int {
	t.Helper()
	iter, err := datastore.CtxMustGetTxn(ctx).Datastore().Iterator(
		ctx, datastore.IterOptions{Prefix: prefix, KeysOnly: true},
	)
	require.NoError(t, err)
	count := 0
	for {
		found, err := iter.Next()
		if err != nil {
			require.NoError(t, errors.Join(err, iter.Close()))
		}
		if !found {
			break
		}
		count++
	}
	require.NoError(t, iter.Close())
	return count
}

func readFullTextTermFrequency(
	t *testing.T,
	ctx context.Context,
	db *DB,
	col client.Collection,
	desc client.IndexDescription,
	term string,
	doc *client.Document,
) (uint64, bool) {
	t.Helper()
	txn, err := db.NewTxn(true)
	require.NoError(t, err)
	defer txn.Discard()
	txnCtx := InitContext(ctx, txn)
	collectionShortID, err := id.GetCollectionShortID(txnCtx, col.Version().CollectionID)
	require.NoError(t, err)
	docShortID, found, err := id.GetDocShortID(txnCtx, collectionShortID, doc.ID().String())
	require.NoError(t, err)
	require.True(t, found)
	epoch, err := getIndexEpoch(txnCtx, col.Version().CollectionID, desc.ID)
	require.NoError(t, err)
	key := keys.NewFullTextPostingKey(collectionShortID, desc.ID, epoch, term, docShortID)
	data, err := datastore.CtxMustGetTxn(txnCtx).Datastore().Get(txnCtx, &key)
	if errors.Is(err, corekv.ErrNotFound) {
		return 0, false
	}
	require.NoError(t, err)
	remaining, frequency, err := encoding.DecodeUvarintAscending(data)
	require.NoError(t, err)
	require.Empty(t, remaining)
	return frequency, true
}

func TestFullTextIndex_BackfillWritesPostingsLengthsAndTotals(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	doc := addUserDoc(t, ctx, col, "the cat sat on the mat")
	addUserDoc(t, ctx, col, "the dog")

	desc := newBM25NameIndex(t, ctx, db, col)
	assert.Equal(t, fullTextState{documents: 2, totalLength: 8, postings: 7, lengths: 2},
		readFullTextState(t, ctx, db, col, desc))
	frequency, found := readFullTextTermFrequency(t, ctx, db, col, desc, "the", doc)
	require.True(t, found)
	assert.Equal(t, uint64(2), frequency)
}

func TestFullTextIndex_NoTokenValueLifecycleWritesNoState(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	doc := addUserDoc(t, ctx, col, "--- a")

	desc := newBM25NameIndex(t, ctx, db, col)
	assert.Equal(t, fullTextState{}, readFullTextState(t, ctx, db, col, desc))

	require.NoError(t, doc.Set(ctx, "name", "quiet afternoon"))
	require.NoError(t, col.UpdateDocument(ctx, doc))
	assert.Equal(t, fullTextState{documents: 1, totalLength: 2, postings: 2, lengths: 1},
		readFullTextState(t, ctx, db, col, desc))

	require.NoError(t, doc.Set(ctx, "name", "--- a"))
	require.NoError(t, col.UpdateDocument(ctx, doc))
	assert.Equal(t, fullTextState{}, readFullTextState(t, ctx, db, col, desc))

	deleted, err := col.DeleteDocument(ctx, doc.ID())
	require.NoError(t, err)
	require.True(t, deleted)
	assert.Equal(t, fullTextState{}, readFullTextState(t, ctx, db, col, desc))
}

func TestFullTextIndex_LiveUpdateAndDeleteMaintainStatistics(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)
	desc := newBM25NameIndex(t, ctx, db, col)
	doc := addUserDoc(t, ctx, col, "the cat sat on the mat")
	assert.Equal(t, fullTextState{documents: 1, totalLength: 6, postings: 5, lengths: 1},
		readFullTextState(t, ctx, db, col, desc))

	require.NoError(t, doc.Set(ctx, "name", "a dog"))
	require.NoError(t, col.UpdateDocument(ctx, doc))
	assert.Equal(t, fullTextState{documents: 1, totalLength: 1, postings: 1, lengths: 1},
		readFullTextState(t, ctx, db, col, desc))

	deleted, err := col.DeleteDocument(ctx, doc.ID())
	require.NoError(t, err)
	require.True(t, deleted)
	assert.Equal(t, fullTextState{}, readFullTextState(t, ctx, db, col, desc))
}

func TestDerivedIndex_TypedCreationValidation(t *testing.T) {
	defaultFullText := func() *client.FullTextIndexDescription {
		return &client.FullTextIndexDescription{
			Algorithm: client.FullTextAlgorithmBM25,
			BM25: &client.BM25Params{
				K1: client.DefaultBM25K1,
				B:  client.DefaultBM25B,
			},
		}
	}
	tests := []struct {
		name    string
		request client.NewIndexRequest
		wantErr error
	}{
		{
			name: "unsupported full-text algorithm",
			request: client.NewIndexRequest{
				Fields:   []client.IndexedFieldDescription{{Name: "name"}},
				FullText: &client.FullTextIndexDescription{Algorithm: "TFIDF"},
			},
			wantErr: NewErrUnsupportedFullTextAlgorithm("TFIDF"),
		},
		{
			name: "negative k1",
			request: client.NewIndexRequest{
				Fields: []client.IndexedFieldDescription{{Name: "name"}},
				FullText: &client.FullTextIndexDescription{
					Algorithm: client.FullTextAlgorithmBM25,
					BM25:      &client.BM25Params{K1: -1, B: 0.75},
				},
			},
			wantErr: NewErrInvalidBM25Parameter("k1", -1),
		},
		{
			name: "nonfinite k1",
			request: client.NewIndexRequest{
				Fields: []client.IndexedFieldDescription{{Name: "name"}},
				FullText: &client.FullTextIndexDescription{
					Algorithm: client.FullTextAlgorithmBM25,
					BM25:      &client.BM25Params{K1: math.Inf(1), B: 0.75},
				},
			},
			wantErr: NewErrInvalidBM25Parameter("k1", math.Inf(1)),
		},
		{
			name: "out-of-range b",
			request: client.NewIndexRequest{
				Fields: []client.IndexedFieldDescription{{Name: "name"}},
				FullText: &client.FullTextIndexDescription{
					Algorithm: client.FullTextAlgorithmBM25,
					BM25:      &client.BM25Params{K1: 1.2, B: 2},
				},
			},
			wantErr: NewErrInvalidBM25Parameter("b", 2),
		},
		{
			name: "nonfinite b",
			request: client.NewIndexRequest{
				Fields: []client.IndexedFieldDescription{{Name: "name"}},
				FullText: &client.FullTextIndexDescription{
					Algorithm: client.FullTextAlgorithmBM25,
					BM25:      &client.BM25Params{K1: 1.2, B: math.NaN()},
				},
			},
			wantErr: NewErrInvalidBM25Parameter("b", math.NaN()),
		},
		{
			name: "descending full-text",
			request: client.NewIndexRequest{
				Fields:   []client.IndexedFieldDescription{{Name: "name", Descending: true}},
				FullText: defaultFullText(),
			},
			wantErr: NewErrStringIndexCannotBeDescending("full-text", "name"),
		},
		{
			name: "descending trigram",
			request: client.NewIndexRequest{
				Fields:  []client.IndexedFieldDescription{{Name: "name", Descending: true}},
				Trigram: &client.TrigramIndexDescription{},
			},
			wantErr: NewErrStringIndexCannotBeDescending("trigram", "name"),
		},
		{
			name: "mutually exclusive kind configs",
			request: client.NewIndexRequest{
				Fields:   []client.IndexedFieldDescription{{Name: "name"}},
				FullText: defaultFullText(),
				Trigram:  &client.TrigramIndexDescription{},
			},
			wantErr: ErrMultipleIndexKindDescriptions,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			_, col := setupUserCollection(t, ctx)
			_, err := col.NewIndex(ctx, test.request)
			require.Error(t, err)
			assert.Equal(t, test.wantErr.Error(), err.Error())
		})
	}
}
