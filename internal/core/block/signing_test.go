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
	"time"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/lock"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/immutable"
)

func TestEnabledSigningFromContext_WithFullIdentity_ReturnsIdentity(t *testing.T) {
	ident, err := identity.Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)

	ctx := context.Background()
	ctx = iIdentity.WithContext(ctx, immutable.Some[identity.Identity](ident))
	ctx = ContextWithEnabledSigning(ctx)

	enabled, fullIdent := EnabledSigningFromContext(ctx)
	require.True(t, enabled)
	require.True(t, fullIdent.HasValue())
	require.NotNil(t, fullIdent.Value().PrivateKey())
}

func TestEnabledSigningFromContext_WithTokenIdentityNilPrivateKey_ReturnsNone(t *testing.T) {
	// Generate an identity and create a bearer token from it.
	ident, err := identity.Generate(crypto.KeyTypeSecp256k1)
	require.NoError(t, err)
	err = ident.UpdateToken(time.Hour, immutable.Some("test"), immutable.None[string]())
	require.NoError(t, err)

	// Parse the token back, this creates a FullIdentity with nil PrivateKey.
	// This is exactly what happens on the server side when a client sends a
	// request with a bearer token (e.g. --no-keyring with --identity).
	tokenIdent, err := identity.FromToken([]byte(ident.BearerToken()))
	require.NoError(t, err)

	ctx := context.Background()
	ctx = iIdentity.WithContext(ctx, immutable.Some[identity.Identity](tokenIdent))
	ctx = ContextWithEnabledSigning(ctx)

	enabled, fullIdent := EnabledSigningFromContext(ctx)
	require.True(t, enabled)

	// If extractFullIdentity returns Some(fullIdent) with nil PrivateKey, that can
	// lead to a panic in signBlock. After the fix, it returns None so signing is skipped.
	require.False(t, fullIdent.HasValue())
}

func TestEnabledSigningFromContext_WithNoIdentity_ReturnsNone(t *testing.T) {
	ctx := context.Background()
	ctx = ContextWithEnabledSigning(ctx)

	enabled, fullIdent := EnabledSigningFromContext(ctx)
	require.True(t, enabled)
	require.False(t, fullIdent.HasValue())
}

func TestEnabledSigningFromContext_WithSigningDisabled_ReturnsFalse(t *testing.T) {
	ctx := context.Background()

	enabled, fullIdent := EnabledSigningFromContext(ctx)
	require.False(t, enabled)
	require.False(t, fullIdent.HasValue())
}

func TestAddDelta_SigningDoesNotChangeBlockCID(t *testing.T) {
	unsignedLink, unsignedBytes, unsignedBlock, unsignedSignatures := addCompositeDeltaForSigningTest(t, false)
	signedLink, signedBytes, signedBlock, signedSignatures := addCompositeDeltaForSigningTest(t, true)

	require.Equal(t, unsignedLink, signedLink)
	require.Equal(t, unsignedBytes, signedBytes)
	require.Nil(t, unsignedBlock.Signature)
	require.Nil(t, signedBlock.Signature)
	require.Empty(t, unsignedSignatures)
	require.Len(t, signedSignatures, 1)
}

func addCompositeDeltaForSigningTest(
	t *testing.T,
	signingEnabled bool,
) (cidlink.Link, []byte, *Block, []StoredSignatureLink) {
	t.Helper()

	rootstore := memory.NewDatastore(context.Background())
	txn := datastore.NewTxnFrom(rootstore, lock.NewLockSet(), 0, false, immutable.None[int]())
	ctx := datastore.CtxSetTxn(context.Background(), txn)

	if signingEnabled {
		ident, err := identity.Generate(crypto.KeyTypeSecp256k1)
		require.NoError(t, err)

		ctx = iIdentity.WithContext(ctx, immutable.Some[identity.Identity](ident))
		ctx = ContextWithEnabledSigning(ctx)
	}

	key := keys.DataStoreKey{}.
		WithCollectionRoot(1).
		WithDocID("docID").
		WithValueFlag()
	composite := crdt.NewDocComposite(txn.Datastore(), "collectionVersionID", key)

	link, blockBytes, err := AddDelta(ctx, composite, composite.Delta())
	require.NoError(t, err)

	rawBlock, err := txn.Blockstore().Get(ctx, link.Cid)
	require.NoError(t, err)

	block, err := GetFromBytes(rawBlock.RawData())
	require.NoError(t, err)

	signatures, err := ListSignatureLinks(ctx, txn.Systemstore(), link.Cid)
	require.NoError(t, err)

	return link, blockBytes, block, signatures
}
