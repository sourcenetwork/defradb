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

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/utils"
)

// VerifySignature verifies the signatures of a block using a public key.
// Returns an error if any signature verification fails.
func (db *DB) VerifySignature(
	ctx context.Context,
	blockCid string,
	pubKey crypto.PublicKey,
	opts ...options.Enumerable[options.VerifySignatureOptions],
) error {
	txn, hadTxn := datastore.CtxTryGetTxn(ctx)

	opt := utils.NewOptions(opts...)

	if err := db.checkNodeAccess(ctx, opt.Identity, acpTypes.NodeVerifySignaturePerm); err != nil {
		return err
	}

	parsedCid, err := cid.Parse(blockCid)
	if err != nil {
		return err
	}

	// If we have a transaction, we will use it to set the blockstore. Otherwise, we will use the db.
	var blockStore *bsadapter.Adapter
	if hadTxn {
		blockStore = &bsadapter.Adapter{Wrapped: datastore.BlockstoreFrom(txn.Rootstore(), db.blockStoreChunkSize)}
	} else {
		blockStore = &bsadapter.Adapter{Wrapped: datastore.BlockstoreFrom(db.rootstore, db.blockStoreChunkSize)}
	}

	linkSys := cidlink.DefaultLinkSystem()
	linkSys.SetReadStorage(blockStore)
	linkSys.TrustedStorage = true

	nd, err := linkSys.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: parsedCid}, coreblock.BlockSchemaPrototype)
	if err != nil {
		return err
	}

	block, err := coreblock.GetFromNode(nd)
	if err != nil {
		return err
	}

	if block.Signature == nil {
		return ErrMissingSignature
	}

	if db.documentACP.HasValue() {
		// Verifying a signature requires read access to the block's document. See
		// [acpDB.CheckDocReadAccess] for the canonical rules (an explicit grant on the document
		// suffices, otherwise a branchable collection also gates on the collection object). We
		// resolve by version id, not docID, so collection-level commits (which have no docID) work.
		versionID := block.Delta.GetCollectionVersionID()
		// Pass the requester's identity so the collection lookup is authorised as them (rather than
		// anonymously) when node acp gates the get-collection operation.
		getColOpts := options.GetCollections().SetGetInactive(true).SetVersionID(versionID)
		if opt.Identity.HasValue() {
			getColOpts = getColOpts.SetIdentity(opt.Identity.Value())
		}
		cols, err := db.GetCollections(ctx, getColOpts)
		if err != nil {
			return err
		}
		if len(cols) == 0 {
			return client.NewErrCollectionNotFoundForCollectionVersion(versionID)
		}

		hasPerm, err := acpDB.CheckDocReadAccess(
			ctx,
			opt.Identity,
			db.nodeACP,
			db.documentACP.Value(),
			cols[0],
			string(block.Delta.GetDocID()),
		)
		if err != nil {
			return err
		}
		if !hasPerm {
			return ErrMissingPermission
		}
	}

	_, err = coreblock.VerifyBlockSignatureWithKey(block, &linkSys, pubKey)
	return err
}
