// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package coreblock

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/lock"
	"github.com/sourcenetwork/defradb/internal/encryption"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/immutable"
)

func TestNewEncryptionBlockSkipsEmptyKey(t *testing.T) {
	require.Nil(t, newEncryptionBlock(nil))
	require.Nil(t, newEncryptionBlock([]byte{}))

	require.Equal(t, &Encryption{Key: []byte("key")}, newEncryptionBlock([]byte("key")))
}

func TestAddDelta_DoesNotEncryptCollectionBlocks(t *testing.T) {
	ctx := context.Background()
	txn := datastore.NewTxnFrom(memory.NewDatastore(ctx), lock.NewLockSet(), 1, false, immutable.None[int]())
	ctx = datastore.CtxSetTxn(ctx, txn)
	ctx = encryption.SetContextConfigFromParams(ctx, true, nil)

	collectionCRDT := crdt.NewCollection("collection-version", keys.NewHeadstoreColKey(1))
	_, rawBlock, err := AddDelta(ctx, collectionCRDT, collectionCRDT.Delta())
	require.NoError(t, err)

	block, err := GetFromBytes(rawBlock)
	require.NoError(t, err)
	require.Nil(t, block.Encryption)
}
