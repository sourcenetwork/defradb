// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package kms

import (
	"context"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kvblockstore "github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/corekv/memory"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

func TestMarshalFetchEncryptionKeyRequest_GeneratesDistinctRequestIDs(t *testing.T) {
	req := &fetchEncryptionKeyRequest{
		Links:              [][]byte{[]byte("encryption-block-link")},
		EphemeralPublicKey: []byte("requester-public-key"),
	}

	firstData, err := marshalFetchEncryptionKeyRequest(req)
	require.NoError(t, err)
	firstRequestID := append([]byte(nil), req.RequestID...)

	secondData, err := marshalFetchEncryptionKeyRequest(req)
	require.NoError(t, err)

	require.Len(t, firstRequestID, 16)
	require.Len(t, req.RequestID, 16)
	assert.NotEqual(t, firstRequestID, req.RequestID)
	assert.NotEqual(t, firstData, secondData)
}

func TestTryHandleFetchEncryptionKeyResponse_UsesReplySenderForAAD(t *testing.T) {
	ctx := context.Background()

	requesterPrivKey, err := crypto.GenerateX25519()
	require.NoError(t, err)
	holderPrivKey, err := crypto.GenerateX25519()
	require.NoError(t, err)

	req := &fetchEncryptionKeyRequest{
		Links:              [][]byte{[]byte("encryption-block-link")},
		EphemeralPublicKey: requesterPrivKey.PublicKey().Bytes(),
	}

	encBlock := &coreblock.Encryption{
		DocID: []byte("doc1"),
		Key:   []byte("doc-key"),
	}
	plainBlock, err := encBlock.Marshal()
	require.NoError(t, err)

	encryptedBlock, err := crypto.EncryptECIES(
		plainBlock,
		requesterPrivKey.PublicKey(),
		crypto.WithAAD(makeAssociatedData(req, "holder-peer")),
		crypto.WithPrivKey(holderPrivKey),
		crypto.WithPubKeyPrepended(false),
	)
	require.NoError(t, err)

	reply := &fetchEncryptionKeyReply{
		Links:              req.Links,
		Blocks:             [][]byte{encryptedBlock},
		EphemeralPublicKey: holderPrivKey.PublicKey().Bytes(),
		Sender:             "holder-peer",
	}
	data, err := cbor.Marshal(reply)
	require.NoError(t, err)

	service := &pubSubService{
		ctx:      ctx,
		encStore: newIPLDEncryptionStorage(datastore.EncstoreFrom(memory.NewDatastore(ctx))),
	}
	items, ok, err := service.tryHandleFetchEncryptionKeyResponse(
		client.PubsubResponse{
			From: "relay-peer",
			Data: data,
		},
		req,
		requesterPrivKey,
		true,
	)

	require.NoError(t, err)
	require.True(t, ok)
	require.Len(t, items, 1)
	assert.Equal(t, req.Links[0], items[0].Link)
	assert.Equal(t, plainBlock, items[0].Block)
}

func TestTryHandleFetchEncryptionKeyResponse_RejectsMismatchedLinksAndBlocks(t *testing.T) {
	ctx := context.Background()

	requesterPrivKey, err := crypto.GenerateX25519()
	require.NoError(t, err)

	req := &fetchEncryptionKeyRequest{
		Links:              [][]byte{[]byte("encryption-block-link")},
		EphemeralPublicKey: requesterPrivKey.PublicKey().Bytes(),
	}

	reply := &fetchEncryptionKeyReply{
		Links:  req.Links,
		Blocks: [][]byte{[]byte("encrypted-block"), []byte("extra-encrypted-block")},
	}
	data, err := cbor.Marshal(reply)
	require.NoError(t, err)

	service := &pubSubService{
		ctx:      ctx,
		encStore: newIPLDEncryptionStorage(datastore.EncstoreFrom(memory.NewDatastore(ctx))),
	}
	items, ok, err := service.tryHandleFetchEncryptionKeyResponse(
		client.PubsubResponse{
			From: "holder-peer",
			Data: data,
		},
		req,
		requesterPrivKey,
		true,
	)

	require.Error(t, err)
	require.False(t, ok)
	assert.Nil(t, items)
}

func TestGetEncryptionKeysLocally_ReturnsOnlyFoundLinks(t *testing.T) {
	ctx := context.Background()
	rootstore := memory.NewDatastore(ctx)
	encstore := datastore.EncstoreFrom(rootstore)
	service := &pubSubService{
		ctx:      ctx,
		encStore: newIPLDEncryptionStorage(encstore),
	}

	foundBlock := &coreblock.Encryption{
		DocID: []byte("doc1"),
		Key:   []byte("doc-key"),
	}
	foundLink := storeEncryptionBlock(t, ctx, encstore, foundBlock)

	missingBlock := &coreblock.Encryption{
		DocID: []byte("doc2"),
		Key:   []byte("other-doc-key"),
	}
	missingLink := storeEncryptionBlock(t, ctx, datastore.EncstoreFrom(memory.NewDatastore(ctx)), missingBlock)

	links, blocks, err := service.getEncryptionKeysLocally(ctx, &fetchEncryptionKeyRequest{
		Links: [][]byte{missingLink, foundLink},
	})

	require.NoError(t, err)
	require.Equal(t, [][]byte{foundLink}, links)
	require.Len(t, blocks, 1)

	foundBlockBytes, err := foundBlock.Marshal()
	require.NoError(t, err)
	assert.Equal(t, foundBlockBytes, blocks[0])
}

func TestTryHandleFetchEncryptionKeyResponse_RejectsUnverifiedBlocks(t *testing.T) {
	ctx := context.Background()

	requesterPrivKey, err := crypto.GenerateX25519()
	require.NoError(t, err)
	holderPrivKey, err := crypto.GenerateX25519()
	require.NoError(t, err)

	req := &fetchEncryptionKeyRequest{
		Links:              [][]byte{[]byte("encryption-block-link")},
		EphemeralPublicKey: requesterPrivKey.PublicKey().Bytes(),
	}

	encBlock := &coreblock.Encryption{
		DocID: []byte("doc1"),
		Key:   []byte("doc-key"),
	}
	plainBlock, err := encBlock.Marshal()
	require.NoError(t, err)

	encryptedBlock, err := crypto.EncryptECIES(
		plainBlock,
		requesterPrivKey.PublicKey(),
		crypto.WithAAD(makeAssociatedData(req, "holder-peer")),
		crypto.WithPrivKey(holderPrivKey),
		crypto.WithPubKeyPrepended(false),
	)
	require.NoError(t, err)

	reply := &fetchEncryptionKeyReply{
		Links:              req.Links,
		Blocks:             [][]byte{encryptedBlock},
		EphemeralPublicKey: holderPrivKey.PublicKey().Bytes(),
	}
	data, err := cbor.Marshal(reply)
	require.NoError(t, err)

	service := &pubSubService{
		ctx:      ctx,
		encStore: newIPLDEncryptionStorage(datastore.EncstoreFrom(memory.NewDatastore(ctx))),
	}

	// service.encStore.computeLink([]byte("encrypted-block"))

	items, ok, err := service.tryHandleFetchEncryptionKeyResponse(
		client.PubsubResponse{
			From: "holder-peer",
			Data: data,
		},
		req,
		requesterPrivKey,
		false,
	)

	require.Error(t, err)
	require.False(t, ok)
	require.Len(t, items, 0)
}

func TestTryHandleFetchEncryptionKeyResponse_AcceptsVerifiedBlocks(t *testing.T) {
	ctx := context.Background()

	service := &pubSubService{
		ctx:      ctx,
		encStore: newIPLDEncryptionStorage(datastore.EncstoreFrom(memory.NewDatastore(ctx))),
	}

	requesterPrivKey, err := crypto.GenerateX25519()
	require.NoError(t, err)
	holderPrivKey, err := crypto.GenerateX25519()
	require.NoError(t, err)

	req := &fetchEncryptionKeyRequest{
		Links:              [][]byte{[]byte("bafyreidq2l2x5zntplo7hebxmj2w6uyw7srv2psrz6pjfgmaaesrb4jauu")},
		EphemeralPublicKey: requesterPrivKey.PublicKey().Bytes(),
	}

	encBlock := &coreblock.Encryption{
		DocID: []byte("doc1"),
		Key:   []byte("doc-key"),
	}
	plainBlock, err := encBlock.Marshal()
	require.NoError(t, err)

	encryptedBlock, err := crypto.EncryptECIES(
		plainBlock,
		requesterPrivKey.PublicKey(),
		crypto.WithAAD(makeAssociatedData(req, "holder-peer")),
		crypto.WithPrivKey(holderPrivKey),
		crypto.WithPubKeyPrepended(false),
	)
	require.NoError(t, err)

	reply := &fetchEncryptionKeyReply{
		Links:              req.Links,
		Blocks:             [][]byte{encryptedBlock},
		EphemeralPublicKey: holderPrivKey.PublicKey().Bytes(),
	}
	data, err := cbor.Marshal(reply)
	require.NoError(t, err)

	// service.encStore.computeLink([]byte("encrypted-block"))

	items, ok, err := service.tryHandleFetchEncryptionKeyResponse(
		client.PubsubResponse{
			From: "holder-peer",
			Data: data,
		},
		req,
		requesterPrivKey,
		false,
	)

	require.NoError(t, err)
	require.True(t, ok)
	require.Len(t, items, 1)
	assert.Equal(t, req.Links[0], items[0].Link)
	assert.Equal(t, plainBlock, items[0].Block)
}

func storeEncryptionBlock(
	t *testing.T,
	ctx context.Context,
	encstore datastore.Blockstore,
	encBlock *coreblock.Encryption,
) []byte {
	t.Helper()

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetWriteStorage(kvblockstore.NewIPLDStore(encstore))

	link, err := lsys.Store(
		linking.LinkContext{Ctx: ctx},
		coreblock.GetLinkPrototype(),
		encBlock.GenerateNode(),
	)
	require.NoError(t, err)

	return []byte(link.Binary())
}
