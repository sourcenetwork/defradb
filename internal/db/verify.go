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
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/datastore"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/db/permission"
)

// VerifySignature verifies the signatures of a block using a public key.
// Returns an error if any signature verification fails.
func (db *DB) VerifySignature(ctx context.Context, blockCid string, pubKey crypto.PublicKey) error {
	parsedCid, err := cid.Parse(blockCid)
	if err != nil {
		return err
	}

	blockStore := &bsadapter.Adapter{Wrapped: datastore.BlockstoreFrom(db.rootstore)}
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
		docID := string(block.Delta.GetDocID())
		collection, err := NewCollectionRetriever(db).RetrieveCollectionFromDocID(ctx, docID)
		if err != nil {
			return err
		}

		hasPerm, err := permission.CheckAccessOfDocOnCollectionWithACP(
			ctx,
			identity.FromContext(ctx),
			db.documentACP.Value(),
			collection,
			acpTypes.DocumentReadPerm,
			docID,
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
