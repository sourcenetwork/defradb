// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/db/lock"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func TestKindForKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want ReadKind
	}{
		{"datastore", "d/1/v/2/3", ReadDoc},
		{"headstore", "h/1/2", ReadHead},
		{"blockstore", "bsomecid", ReadBlock},
		{"systemstore", "s/collection/name/Foo", ReadSystem},
		{"sequence", "s/seq/doc", ReadSequence},
		{"peerstore", "p/replicator/1", ReadPeer},
		{"encstore", "esomecid", ReadEnc},
		{"unknown", "?/1", 0},
		{"empty", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, kindForKey([]byte(tt.key)))
		})
	}
}

func TestReadKindString(t *testing.T) {
	require.Equal(t, "none", ReadKind(0).String())
	require.Equal(t, "doc", ReadDoc.String())
	require.Equal(t, "doc|head|rangeIter", (ReadDoc | ReadHead | ReadRangeIterator).String())
}

// The recorder has to see the store discriminator that namespace.Wrap prepends, so these
// go through a real transaction rather than calling the recorder directly.
func TestTxnRecordsReadKinds(t *testing.T) {
	ctx := context.Background()
	txn := NewTxnFrom(memory.NewDatastore(ctx), lock.NewLockSet(), 0, false, immutable.None[int]())
	defer txn.Discard()
	ctx = CtxSetTxn(ctx, txn)

	require.Equal(t, ReadKind(0), txn.ReadKinds().Kinds())

	key := keys.DataStoreKey{CollectionShortID: 1, InstanceType: keys.ValueKey, DocShortID: 2, FieldID: "3"}
	_, err := txn.Datastore().Get(ctx, key)
	require.ErrorIs(t, err, corekv.ErrNotFound)
	require.Equal(t, ReadDoc, txn.ReadKinds().Kinds())

	_, err = txn.Headstore().Get(ctx, []byte("/1/2"))
	require.ErrorIs(t, err, corekv.ErrNotFound)
	require.Equal(t, ReadDoc|ReadHead, txn.ReadKinds().Kinds())
}

// A start/end iterator is the shape whose bounds check reads one key past the requested
// range. Distinguishing it from a prefix iterator is the point of the two iterator bits.
func TestTxnRecordsIteratorShape(t *testing.T) {
	ctx := context.Background()

	prefix := keys.DataStoreKey{CollectionShortID: 1, InstanceType: keys.ValueKey, DocShortID: 2}

	rangeTxn := NewTxnFrom(memory.NewDatastore(ctx), lock.NewLockSet(), 0, false, immutable.None[int]())
	defer rangeTxn.Discard()
	iter, err := rangeTxn.Datastore().Iterator(
		CtxSetTxn(ctx, rangeTxn),
		IterOptions{Start: prefix, End: prefix.PrefixEnd()},
	)
	require.NoError(t, err)
	require.NoError(t, iter.Close())
	require.Equal(t, ReadDoc|ReadRangeIterator, rangeTxn.ReadKinds().Kinds())

	prefixTxn := NewTxnFrom(memory.NewDatastore(ctx), lock.NewLockSet(), 0, false, immutable.None[int]())
	defer prefixTxn.Discard()
	iter, err = prefixTxn.Datastore().Iterator(CtxSetTxn(ctx, prefixTxn), IterOptions{Prefix: prefix})
	require.NoError(t, err)
	require.NoError(t, iter.Close())
	require.Equal(t, ReadDoc|ReadPrefixIterator, prefixTxn.ReadKinds().Kinds())
}
