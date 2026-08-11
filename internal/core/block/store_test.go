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

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/lock"
	"github.com/sourcenetwork/defradb/internal/encryption"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
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

func TestAddDelta_WithBatchCollector_CollectsCompositeCIDsOnly(t *testing.T) {
	ctx := context.Background()
	txn := datastore.NewTxnFrom(memory.NewDatastore(ctx), lock.NewLockSet(), 1, false, immutable.None[int]())
	ctx = datastore.CtxSetTxn(ctx, txn)
	store := txn.Datastore()

	ident, err := identity.Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)
	ctx = iIdentity.WithContext(ctx, immutable.Some[identity.Identity](ident))
	ctx = ContextWithEnabledSigning(ctx)

	collector := NewBatchCIDCollector()
	ctx = ContextWithBatchSigning(ctx, collector)

	key := keys.DataStoreKey{CollectionShortID: 1, DocShortID: 42}

	// A composite block is a document root, so its CID is collected. In batch mode
	// the block is also left unsigned.
	composite := crdt.NewDocComposite(store, "collection-version", key)
	link, rawBlock, err := AddDelta(ctx, composite, composite.Delta())
	require.NoError(t, err)
	block, err := GetFromBytes(rawBlock)
	require.NoError(t, err)
	require.Nil(t, block.Signature)
	require.Equal(t, []cid.Cid{link.Cid}, collector.GetCIDs())

	// A field block is not a document root, so its CID is not collected. Otherwise a
	// field value shared across documents would duplicate CIDs.
	fieldKey := key
	fieldKey.FieldID = "1"
	lww := crdt.NewLWW(store, "collection-version", fieldKey, "name")
	lwwDelta, err := lww.Delta(ctx, crdt.NewDocField("name", client.NewFieldValue(client.LWW_REGISTER, client.NewNormalString("Alice"))))
	require.NoError(t, err)
	_, _, err = AddDelta(ctx, lww, lwwDelta)
	require.NoError(t, err)
	require.Equal(t, []cid.Cid{link.Cid}, collector.GetCIDs())
}

func TestAddDelta_WithoutBatchCollector_SignsBlock(t *testing.T) {
	ctx := context.Background()
	txn := datastore.NewTxnFrom(memory.NewDatastore(ctx), lock.NewLockSet(), 1, false, immutable.None[int]())
	ctx = datastore.CtxSetTxn(ctx, txn)

	ident, err := identity.Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)
	ctx = iIdentity.WithContext(ctx, immutable.Some[identity.Identity](ident))
	ctx = ContextWithEnabledSigning(ctx)

	collectionCRDT := crdt.NewCollection("collection-version", keys.NewHeadstoreColKey(1))
	_, rawBlock, err := AddDelta(ctx, collectionCRDT, collectionCRDT.Delta())
	require.NoError(t, err)

	block, err := GetFromBytes(rawBlock)
	require.NoError(t, err)
	require.NotNil(t, block.Signature)
}

// A signature block is not among a block's links, so it has to be cleared alongside the
// block's own marker. Left marked, a signature fetched over P2P stays sweep-eligible for
// the life of the store.
func TestProcessBlock_ClearsSignatureMarker(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)

	// Store the signature block the way an inbound one arrives. The P2P blockstore is what
	// stamps the to-merge marker; a locally written block never carries one.
	sigBlock := blocks.NewBlock([]byte("signature"))
	p2pStore := datastore.P2PBlockstoreFrom(store, immutable.None[int]())
	require.NoError(t, p2pStore.Put(ctx, sigBlock))

	merged, err := p2pStore.IsMerged(ctx, sigBlock.Cid())
	require.NoError(t, err)
	require.False(t, merged, "guard: the signature block must start out marked to-merge")

	txn := datastore.NewTxnFrom(store, lock.NewLockSet(), 1, false, immutable.None[int]())
	txnCtx := datastore.CtxSetTxn(ctx, txn)

	collectionCRDT := crdt.NewCollection("collection-version", keys.NewHeadstoreColKey(1))
	block := New(crdt.NewCRDT(collectionCRDT.Delta()), nil)
	block.Signature = &cidlink.Link{Cid: sigBlock.Cid()}

	link, err := block.GenerateLink()
	require.NoError(t, err)

	require.NoError(t, ProcessBlock(txnCtx, collectionCRDT, block, link))
	require.NoError(t, txn.Commit())

	merged, err = p2pStore.IsMerged(ctx, sigBlock.Cid())
	require.NoError(t, err)
	require.True(t, merged, "merging a signed block must clear its signature block's marker")
}
