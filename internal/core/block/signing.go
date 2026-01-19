// Copyright 2025 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/immutable"
)

type enabledSigningContextKey struct{}

// ContextWithEnabledSigning returns a context with block signing enabled.
func ContextWithEnabledSigning(ctx context.Context) context.Context {
	return context.WithValue(ctx, enabledSigningContextKey{}, true)
}

// EnabledSigningFromContext returns true if block signing is enabled in the context.
func EnabledSigningFromContext(ctx context.Context) (bool, immutable.Option[identity.FullIdentity]) {
	val := ctx.Value(enabledSigningContextKey{})
	if val == nil {
		return false, immutable.None[identity.FullIdentity]()
	}

	return val.(bool), extractFullIdentity(ctx) //nolint:forcetypeassert
}

func extractFullIdentity(ctx context.Context) immutable.Option[identity.FullIdentity] {
	ident := identity.FromContext(ctx)
	if !ident.HasValue() {
		return immutable.None[identity.FullIdentity]()
	}

	fullIdent, ok := ident.Value().(identity.FullIdentity)
	if !ok {
		return immutable.None[identity.FullIdentity]()
	}

	return immutable.Some(fullIdent)
}

func signBlock(
	ctx context.Context,
	blockstore datastore.Blockstore,
	block *Block,
	ident identity.FullIdentity,
) error {
	// We sign only the first field blocks just to add entropy and prevent any collisions.
	// The integrity of the field data is guaranteed by signatures of the parent composite blocks.
	if block.Delta.IsField() && block.Delta.GetPriority() > 1 {
		return nil
	}

	blockBytes, err := block.Marshal()
	if err != nil {
		return err
	}

	var sigType string

	switch ident.PrivateKey().Type() {
	case crypto.KeyTypeSecp256k1:
		sigType = SignatureTypeECDSA256K
	case crypto.KeyTypeEd25519:
		sigType = SignatureTypeEd25519
	default:
		return NewErrUnsupportedKeyForSigning(ident.PrivateKey().Type())
	}

	sigBytes, err := ident.PrivateKey().Sign(blockBytes)
	if err != nil {
		return err
	}

	sig := &Signature{
		Header: SignatureHeader{
			Type:     sigType,
			Identity: []byte(ident.PublicKey().String()),
		},
		Value: sigBytes,
	}

	sigBlockLink, err := putBlock(ctx, blockstore, sig)
	if err != nil {
		return err
	}

	block.Signature = &sigBlockLink

	return nil
}
