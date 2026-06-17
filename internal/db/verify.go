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
	"github.com/sourcenetwork/defradb/internal/db/id"
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
		getCollectionsOpt := options.GetCollections().SetVersionID(block.Delta.GetCollectionVersionID())
		if opt.Identity.HasValue() {
			getCollectionsOpt = getCollectionsOpt.SetIdentity(opt.Identity.Value())
		}
		collections, err := db.GetCollections(ctx, getCollectionsOpt)
		if err != nil {
			return err
		}
		if len(collections) == 0 {
			return ErrMissingPermission
		}
		collection := collections[0]

		docIDs, err := db.publicDocIDsForSignatureBlock(ctx, parsedCid, block, collection)
		if err != nil {
			return err
		}

		var hasPerm bool
		for _, docID := range docIDs {
			hasPerm, err = acpDB.CheckAccessOfDocOnCollectionWithACP(
				ctx,
				opt.Identity,
				db.nodeACP,
				db.documentACP.Value(),
				collection,
				acpTypes.DocumentReadPerm,
				docID,
			)
			if err != nil {
				return err
			}
			if hasPerm {
				break
			}
		}

		if !hasPerm {
			return ErrMissingPermission
		}
	}

	_, err = coreblock.VerifyBlockSignatureWithKey(block, &linkSys, pubKey)
	return err
}

// publicDocIDsForSignatureBlock resolves the public DocIDs that ACP may check for a signed block.
func (db *DB) publicDocIDsForSignatureBlock(
	ctx context.Context,
	blockCID cid.Cid,
	block *coreblock.Block,
	collection client.Collection,
) ([]string, error) {
	if block.Delta.IsCollection() {
		return []string{""}, nil
	}

	shortID, err := id.GetUncachedShortCollectionID(
		ctx,
		collection.Version().CollectionID,
		datastore.SystemstoreFrom(db.rootstore),
	)
	if err != nil {
		return nil, err
	}

	docIDs, err := id.GetDocIDsForBlockFromStore(
		ctx,
		datastore.SystemstoreFrom(db.rootstore),
		shortID,
		blockCID,
	)
	if err != nil {
		return nil, err
	}
	if len(docIDs) > 0 {
		return docIDs, nil
	}

	if block.Delta.IsComposite() && len(block.Heads) == 0 {
		return []string{client.NewDocIDV0(blockCID).String()}, nil
	}
	return []string{""}, nil
}
