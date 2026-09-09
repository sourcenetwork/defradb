// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package blockowner

import (
	"context"
	"fmt"
	"testing"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/memory"

	"github.com/sourcenetwork/defradb/internal/keys"
)

func TestHasOwners(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)

	fieldCID := blocks.NewBlock([]byte("field value")).Cid()

	has, err := HasOwners(ctx, store, fieldCID)
	require.NoError(t, err)
	require.False(t, has, "a block with no recorded owners has none")

	has, err = HasOwners(ctx, store, cid.Undef)
	require.NoError(t, err)
	require.False(t, has, "an undefined CID has no owners")

	setOwner(t, ctx, store, fieldCID, "bae-doc-one")
	setOwner(t, ctx, store, fieldCID, "bae-doc-two")

	has, err = HasOwners(ctx, store, fieldCID)
	require.NoError(t, err)
	require.True(t, has)

	// One of two owners removed: the block is still owned.
	deleteOwner(t, ctx, store, fieldCID, "bae-doc-one")
	has, err = HasOwners(ctx, store, fieldCID)
	require.NoError(t, err)
	require.True(t, has)

	deleteOwner(t, ctx, store, fieldCID, "bae-doc-two")
	has, err = HasOwners(ctx, store, fieldCID)
	require.NoError(t, err)
	require.False(t, has)
}

func TestDocIDs(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)

	fieldCID := blocks.NewBlock([]byte("field value")).Cid()

	docIDs, err := DocIDs(ctx, store, cid.Undef)
	require.NoError(t, err)
	require.Empty(t, docIDs, "an undefined CID has no owners")

	// A block CID can be owned by more than one document.
	setOwner(t, ctx, store, fieldCID, "bae-doc-one")
	setOwner(t, ctx, store, fieldCID, "bae-doc-two")
	docIDs, err = DocIDs(ctx, store, fieldCID)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"bae-doc-one", "bae-doc-two"}, docIDs)

	deleteOwner(t, ctx, store, fieldCID, "bae-doc-one")
	docIDs, err = DocIDs(ctx, store, fieldCID)
	require.NoError(t, err)
	require.Equal(t, []string{"bae-doc-two"}, docIDs)

	deleteOwner(t, ctx, store, fieldCID, "bae-doc-two")
	docIDs, err = DocIDs(ctx, store, fieldCID)
	require.NoError(t, err)
	require.Empty(t, docIDs)
}

// HasOwners reports ownership after reading a single key regardless of how many
// documents own the block, unlike DocIDs which reads every owner.
func TestHasOwnersStopsAtFirstOwner(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)

	fieldCID := blocks.NewBlock([]byte("shared field value")).Cid()
	const owners = 500
	for i := range owners {
		setOwner(t, ctx, store, fieldCID, fmt.Sprintf("bae-doc-%d", i))
	}

	var hasReads int
	has, err := HasOwners(ctx, countingReader{store, &hasReads}, fieldCID)
	require.NoError(t, err)
	require.True(t, has)
	require.Equal(t, 1, hasReads, "HasOwners must stop at the first owner")

	var listReads int
	all, err := DocIDs(ctx, countingReader{store, &listReads}, fieldCID)
	require.NoError(t, err)
	require.Len(t, all, owners)
	require.Greater(t, listReads, owners, "DocIDs reads every owner")
}

// A block whose only owner edge is excluded reports no owner, while the same block
// reports an owner when nothing is excluded.
func TestHasOwnersExceptExcludedSoleOwner(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)

	fieldCID := blocks.NewBlock([]byte("field value")).Cid()
	setOwner(t, ctx, store, fieldCID, "bae-doc-one")

	has, err := HasOwnersExcept(ctx, store, fieldCID, nil)
	require.NoError(t, err)
	require.True(t, has, "with nothing excluded the block has an owner")

	excluded := map[string]struct{}{ownerKey(fieldCID, "bae-doc-one"): {}}
	has, err = HasOwnersExcept(ctx, store, fieldCID, excluded)
	require.NoError(t, err)
	require.False(t, has, "excluding the only owner edge reports no owner")
}

// A block shared by two documents still reports an owner while any of its edges is
// unexcluded, and reports none only once all are.
func TestHasOwnersExceptKeepsUnexcludedOwner(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)

	fieldCID := blocks.NewBlock([]byte("shared field value")).Cid()
	setOwner(t, ctx, store, fieldCID, "bae-doc-one")
	setOwner(t, ctx, store, fieldCID, "bae-doc-two")

	excluded := map[string]struct{}{ownerKey(fieldCID, "bae-doc-one"): {}}
	has, err := HasOwnersExcept(ctx, store, fieldCID, excluded)
	require.NoError(t, err)
	require.True(t, has, "one of two owners excluded: the block is still owned")

	excluded[ownerKey(fieldCID, "bae-doc-two")] = struct{}{}
	has, err = HasOwnersExcept(ctx, store, fieldCID, excluded)
	require.NoError(t, err)
	require.False(t, has, "both owners excluded: no owner remains")
}

// The scan skips excluded owners but stops at the first owner that is not excluded,
// rather than reading the whole owner set.
func TestHasOwnersExceptStopsAtFirstUnexcludedOwner(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)

	fieldCID := blocks.NewBlock([]byte("shared field value")).Cid()
	const owners = 500
	for i := range owners {
		setOwner(t, ctx, store, fieldCID, fmt.Sprintf("bae-doc-%03d", i))
	}

	// Owner edges iterate in key order, so excluding the first makes the scan skip it and
	// stop at the second: two advances, not one and not the whole set.
	excluded := map[string]struct{}{ownerKey(fieldCID, "bae-doc-000"): {}}
	var reads int
	has, err := HasOwnersExcept(ctx, countingReader{store, &reads}, fieldCID, excluded)
	require.NoError(t, err)
	require.True(t, has)
	require.Equal(t, 2, reads, "scan skips the excluded owner and stops at the next")
}

// ownerKey builds the owner-edge key for (blockCID, docID), the same bytes the owner-edge
// iterator yields and that HasOwnersExcept expects its excluded set to hold.
func ownerKey(blockCID cid.Cid, docID string) string {
	return string(keys.NewBlockCIDToDocIDKey(blockCID.String(), docID).Bytes())
}

func setOwner(t *testing.T, ctx context.Context, store corekv.Writer, blockCID cid.Cid, docID string) {
	t.Helper()
	require.NoError(t, store.Set(ctx, []byte(ownerKey(blockCID, docID)), []byte{}))
}

func deleteOwner(t *testing.T, ctx context.Context, store corekv.Writer, blockCID cid.Cid, docID string) {
	t.Helper()
	require.NoError(t, store.Delete(ctx, []byte(ownerKey(blockCID, docID))))
}

// countingReader wraps a corekv.Reader and tallies iterator advances so a test can assert
// how much of a key range a function scans.
type countingReader struct {
	corekv.Reader
	nextCalls *int
}

func (r countingReader) Iterator(ctx context.Context, opts corekv.IterOptions) (corekv.Iterator, error) {
	iter, err := r.Reader.Iterator(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &countingIterator{Iterator: iter, nextCalls: r.nextCalls}, nil
}

type countingIterator struct {
	corekv.Iterator
	nextCalls *int
}

func (it *countingIterator) Next() (bool, error) {
	*it.nextCalls++
	return it.Iterator.Next()
}
