// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"context"

	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/immutable"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/encryption"
)

func makeLinkSystem(blockService blockstore.IPLDStore) linking.LinkSystem {
	linkSys := cidlink.DefaultLinkSystem()
	linkSys.SetWriteStorage(blockService)
	linkSys.SetReadStorage(blockService)
	linkSys.TrustedStorage = true

	return linkSys
}

// syncDAG synchronizes the DAG starting with the given block
// using the blockservice to fetch remote blocks.
//
// This process walks the entire DAG until the issue below is resolved.
// https://github.com/sourcenetwork/defradb/issues/2722
func (p *P2P) syncDAG(ctx context.Context, block *coreblock.Block) error {
	sessionCtx, cancelSession := context.WithCancel(ctx)
	defer cancelSession()

	// use a session to make remote fetches more efficient
	sessionCtx = p.host.ContextWithSession(sessionCtx)

	linkSystem := makeLinkSystem(p.host.IPLDStore())

	// Store the block in the DAG store
	_, err := linkSystem.Store(linking.LinkContext{Ctx: sessionCtx}, coreblock.GetLinkPrototype(), block.GenerateNode())
	if err != nil {
		return NewErrStoreBlockDAGSync(err)
	}

	return p.loadBlockLinks(sessionCtx, &linkSystem, block)
}

// loadBlockLinks loads the links of a block recursively.
//
// The function returns immediately on the first error encountered.
func (p *P2P) loadBlockLinks(ctx context.Context, linkSys *linking.LinkSystem, block *coreblock.Block) error {
	link, err := block.GenerateLink()
	if err != nil {
		return NewErrGenerateBlockLink(err)
	}
	bstore := datastore.BlockstoreFrom(p.db.Rootstore(), immutable.None[int]())
	merged, err := bstore.IsMerged(ctx, link.Cid)
	if err != nil {
		return NewErrCheckBlockMerged(err)
	}
	if merged {
		return nil
	}

	// TODO: this part is not tested yet because there is not easy way of doing it at the moment.
	// https://github.com/sourcenetwork/defradb/issues/3525
	if block.Signature != nil {
		// we deliberately ignore the first returned value, which indicates whether the signature
		// the block was actually verified or not, because we don't handle it any different here.
		// But we want to keep the API of VerifyBlockSignature explicit about the results.
		_, err := coreblock.VerifyBlockSignature(block, linkSys)
		if err != nil {
			return NewErrVerifyBlockSig(err)
		}
	}

	var encResults *encryption.Results
	if block.IsEncrypted() {
		results, err := p.kms.GetKeys(ctx, *block.Encryption)
		if err != nil {
			return NewErrGetEncKeysForBlock(err)
		}
		encResults = results
	}

	for _, lnk := range block.AllLinks() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		ctxWithTimeout, cancel := context.WithTimeout(ctx, p.syncBlockLinkTimeout)
		nd, err := linkSys.Load(linking.LinkContext{Ctx: ctxWithTimeout}, lnk, coreblock.BlockSchemaPrototype)
		cancel()

		if err != nil {
			return NewErrLoadLinkedBlock(err)
		}

		linkBlock, err := coreblock.GetFromNode(nd)
		if err != nil {
			return NewErrDecodeLinkedBlock(err)
		}

		err = p.loadBlockLinks(ctx, linkSys, linkBlock)
		if err != nil {
			return NewErrProcessLinkedBlock(err)
		}
	}

	if encResults != nil {
		for res := range encResults.Get() {
			if res.Error != nil {
				return NewErrRetrieveEncKey(res.Error)
			}
		}
	}

	return nil
}
