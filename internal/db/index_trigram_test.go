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
)

func TestTrigrams(t *testing.T) {
	tests := []struct {
		name     string
		val      string
		expected []string
	}{
		{
			name:     "empty string",
			val:      "",
			expected: nil,
		},
		{
			name:     "shorter than three bytes",
			val:      "hi",
			expected: nil,
		},
		{
			name:     "exactly three bytes",
			val:      "abc",
			expected: []string{"abc"},
		},
		{
			name:     "overlapping windows",
			val:      "abcd",
			expected: []string{"abc", "bcd"},
		},
		{
			name:     "lowercased",
			val:      "ABCD",
			expected: []string{"abc", "bcd"},
		},
		{
			name:     "repeated trigram appears once",
			val:      "banana",
			expected: []string{"ban", "ana", "nan"},
		},
		{
			// The windows are bytes, so a two-byte rune is split across them. The candidates
			// this produces are still correct, since the pattern match is re-run afterwards.
			name:     "unicode is windowed by byte",
			val:      "héllo",
			expected: []string{"h\xc3\xa9", "\xc3\xa9l", "\xa9ll", "llo"},
		},
		{
			// A single three-byte rune is exactly one trigram.
			name:     "three byte rune",
			val:      "€",
			expected: []string{"\xe2\x82\xac"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, trigrams(test.val))
		})
	}
}

// newTrigramNameIndex creates a trigram index on "name" and drains the build, mirroring
// newNameIndex for the ordered key index.
func newTrigramNameIndex(
	t *testing.T,
	ctx context.Context,
	db *DB,
	col client.Collection,
) (client.IndexDescription, error) {
	t.Helper()
	desc, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
		Kind:   client.IndexKindTrigram,
	})
	if err != nil {
		return desc, err
	}
	db.indexBuildWorker.drainSync(ctx)
	return desc, nil
}

// A backfill must write one entry per distinct trigram of each document, and nothing at all
// for a value too short to have one.
func TestTrigramIndex_Backfill_WritesOneEntryPerDistinctTrigram(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	// "banana" has trigrams ban, ana, nan, ana - three distinct.
	addUserDoc(t, ctx, col, "banana")
	// Too short to have any trigram, so it gets no entry.
	addUserDoc(t, ctx, col, "hi")

	desc, err := newTrigramNameIndex(t, ctx, db, col)
	require.NoError(t, err)

	shortID := getCollectionShortID(t, ctx, db, col.Version().CollectionID)
	assert.Equal(t, 3, countIndexEntries(t, ctx, db, shortID, desc.ID))
}

// A live write into an existing trigram index must write the same entries the backfill would,
// and an update must leave only the new value's entries behind.
func TestTrigramIndex_LiveWriteAndUpdate_MaintainEntries(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	desc, err := newTrigramNameIndex(t, ctx, db, col)
	require.NoError(t, err)

	shortID := getCollectionShortID(t, ctx, db, col.Version().CollectionID)
	assert.Equal(t, 0, countIndexEntries(t, ctx, db, shortID, desc.ID))

	doc := addUserDoc(t, ctx, col, "banana")
	assert.Equal(t, 3, countIndexEntries(t, ctx, db, shortID, desc.ID))

	// "peach" has three distinct trigrams: pea, eac, ach.
	require.NoError(t, doc.Set(ctx, "name", "peach"))
	require.NoError(t, col.UpdateDocument(ctx, doc))
	assert.Equal(t, 3, countIndexEntries(t, ctx, db, shortID, desc.ID))

	require.NoError(t, doc.Set(ctx, "name", "hi"))
	require.NoError(t, col.UpdateDocument(ctx, doc))
	assert.Equal(t, 0, countIndexEntries(t, ctx, db, shortID, desc.ID))
}

func TestTrigramIndex_OnDocumentDelete_RemovesEntries(t *testing.T) {
	ctx := context.Background()
	db, col := setupUserCollection(t, ctx)

	desc, err := newTrigramNameIndex(t, ctx, db, col)
	require.NoError(t, err)

	doc := addUserDoc(t, ctx, col, "banana")
	shortID := getCollectionShortID(t, ctx, db, col.Version().CollectionID)
	require.Equal(t, 3, countIndexEntries(t, ctx, db, shortID, desc.ID))

	deleted, err := col.DeleteDocument(ctx, doc.ID())
	require.NoError(t, err)
	require.True(t, deleted)

	assert.Equal(t, 0, countIndexEntries(t, ctx, db, shortID, desc.ID))
}

func TestTrigramIndex_WithUnique_ReturnsError(t *testing.T) {
	ctx := context.Background()
	_, col := setupUserCollection(t, ctx)

	_, err := col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}},
		Kind:   client.IndexKindTrigram,
		Unique: true,
	})
	require.ErrorIs(t, err, ErrTrigramIndexUnique)
}

func TestTrigramIndex_WithMoreThanOneField_ReturnsError(t *testing.T) {
	ctx := context.Background()
	db := newBadgerDBNoIndexWorker(t, ctx)
	_, err := db.AddCollection(ctx, `type User { name: String, alias: String }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	_, err = col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "name"}, {Name: "alias"}},
		Kind:   client.IndexKindTrigram,
	})
	require.ErrorIs(t, err, ErrTrigramIndexNotSingleField)
}

func TestTrigramIndex_OnNonStringField_ReturnsError(t *testing.T) {
	ctx := context.Background()
	db := newBadgerDBNoIndexWorker(t, ctx)
	_, err := db.AddCollection(ctx, `type User { name: String, age: Int }`)
	require.NoError(t, err)
	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	_, err = col.NewIndex(ctx, client.NewIndexRequest{
		Fields: []client.IndexedFieldDescription{{Name: "age"}},
		Kind:   client.IndexKindTrigram,
	})
	require.ErrorIs(t, err, NewErrTrigramIndexFieldNotString("age", "Int"))
	require.ErrorContains(t, err, "age")
	require.ErrorContains(t, err, "Int")
}
