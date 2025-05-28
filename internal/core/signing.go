// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package core

import (
	"context"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/crypto"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
)

type enabledSigningContextKey struct{}

// ContextWithEnabledSigning returns a context with block signing enabled.
func ContextWithEnabledSigning(ctx context.Context) context.Context {
	return context.WithValue(ctx, enabledSigningContextKey{}, true)
}

// EnabledSigningFromContext returns true if block signing is enabled in the context.
func EnabledSigningFromContext(ctx context.Context) bool {
	val := ctx.Value(enabledSigningContextKey{})
	if val == nil {
		return false
	}
	return val.(bool) //nolint:forcetypeassert
}

func signBlock(
	ctx context.Context,
	blockstore datastore.Blockstore,
	block *coreblock.Block,
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

	ident := identity.FromContext(ctx)
	if !ident.HasValue() {
		return nil
	}

	// Check if the identity is a FullIdentity (has private key)
	fullIdent, ok := ident.Value().(identity.FullIdentity)
	if !ok {
		return nil
	}

	var sigType string

	switch fullIdent.PrivateKey().Type() {
	case crypto.KeyTypeSecp256k1:
		sigType = coreblock.SignatureTypeECDSA256K
	case crypto.KeyTypeEd25519:
		sigType = coreblock.SignatureTypeEd25519
	default:
		return coreblock.NewErrUnsupportedKeyForSigning(fullIdent.PrivateKey().Type())
	}

	sigBytes, err := fullIdent.PrivateKey().Sign(blockBytes)
	if err != nil {
		return err
	}

	sig := &coreblock.Signature{
		Header: coreblock.SignatureHeader{
			Type:     sigType,
			Identity: []byte(fullIdent.PublicKey().String()),
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
