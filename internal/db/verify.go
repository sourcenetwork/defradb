// Copyright 2025 Democratized Data Foundation
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

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/bsadapter"

	"github.com/sourcenetwork/defradb/acp/identity"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

// VerifyBlock verifies the signatures of a block.
// The context should carry the identity which will be used to verify the signatures.
// Returns an error if any signature verification fails.
func (db *DB) VerifyBlock(ctx context.Context, blockCID string) error {
	parsedCID, err := cid.Parse(blockCID)
	if err != nil {
		return err
	}

	blockStore := &bsadapter.Adapter{Wrapped: db.Blockstore()}
	linkSys := cidlink.DefaultLinkSystem()
	linkSys.SetReadStorage(blockStore)
	linkSys.TrustedStorage = true

	nd, err := linkSys.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: parsedCID}, coreblock.BlockSchemaPrototype)
	if err != nil {
		return err
	}

	block, err := coreblock.GetFromNode(nd)
	if err != nil {
		return err
	}

	return verifyIdentityAndBlock(ctx, &linkSys, block)
}

func verifyIdentityAndBlock(ctx context.Context, linkSys *linking.LinkSystem, block *coreblock.Block) error {
	if block.Signature == nil {
		return ErrMissingSignature
	}

	ident := identity.FromContext(ctx)
	if !ident.HasValue() {
		return ErrNoIdentityInContext
	}

	nd, err := linkSys.Load(linking.LinkContext{Ctx: ctx}, *block.Signature, coreblock.SignatureSchemaPrototype)
	if err != nil {
		return coreblock.NewErrCouldNotLoadSignatureBlock(err)
	}

	sigBlock, err := coreblock.GetSignatureBlockFromNode(nd)
	if err != nil {
		return coreblock.NewErrCouldNotLoadSignatureBlock(err)
	}

	identityStr := string(sigBlock.Header.Identity)
	if identityStr != ident.Value().PublicKey.String() {
		return NewErrSignatureIdentityMismatch(ident.Value().PublicKey.String())
	}

	_, err = coreblock.VerifyBlockSignature(block, linkSys)
	return err
}
