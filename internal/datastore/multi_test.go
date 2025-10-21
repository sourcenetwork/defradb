// Copyright 2025 Democratized Data Foundation
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
	"strings"
	"testing"

	blocks "github.com/ipfs/go-block-format"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/memory"
)

func TestMultistore_HumanReadableKeys_ShouldSucceed(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)

	ms := NewMultistore(rootstore)

	err := ms.Blockstore().Put(ctx, blocks.NewBlock([]byte("123")))
	require.NoError(t, err)
	err = P2PBlockstoreFrom(rootstore).Put(ctx, blocks.NewBlock([]byte("1234")))
	require.NoError(t, err)
	err = ms.Datastore().Set(ctx, []byte("/123"), []byte("123"))
	require.NoError(t, err)
	err = ms.Encstore().Put(ctx, blocks.NewBlock([]byte("123")))
	require.NoError(t, err)
	err = ms.Headstore().Set(ctx, []byte("/123"), []byte("123"))
	require.NoError(t, err)
	err = ms.Peerstore().Set(ctx, []byte("/123"), []byte("123"))
	require.NoError(t, err)
	err = ms.Systemstore().Set(ctx, []byte("/123"), []byte("123"))
	require.NoError(t, err)

	iter, err := rootstore.Iterator(ctx, corekv.IterOptions{KeysOnly: true})
	require.NoError(t, err)

	expectedKVs := []struct {
		key string
	}{
		{
			key: "blocks/",
		},
		{
			key: "blocks/",
		},
		{
			key: "blocks/to_merge/",
		},
		{
			key: "data/",
		},
		{
			key: "encryption",
		},
		{
			key: "heads/",
		},
		{
			key: "peers/",
		},
		{
			key: "system/",
		},
	}

	expectedIndex := 0
	for {
		hasNext, err := iter.Next()
		require.NoError(t, err)

		if !hasNext {
			break
		}

		key, err := HumanReadableKey(iter.Key())
		require.NoError(t, err)

		require.True(t, strings.HasPrefix(key, expectedKVs[expectedIndex].key), key, expectedKVs[expectedIndex].key)

		expectedIndex++
	}

	iter.Close()
}
