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
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
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

	return p.loadBlockLinks(sessionCtx, &linkSystem, block)
}

// loadBlockLinks loads the links of a block iteratively.
//
// The function returns immediately on the first error encountered.
func (p *P2P) loadBlockLinks(
	ctx context.Context,
	linkSys *linking.LinkSystem,
	block *coreblock.Block,
) error {
	stack := make([]*coreblock.Block, 0, 64)
	visited := make(map[string]struct{})

	processBlock := func(current *coreblock.Block) error {
		// TODO: this part is not tested yet because there is not easy way of doing it at the moment.
		// https://github.com/sourcenetwork/defradb/issues/3525
		link, err := current.GenerateLink()
		if err != nil {
			return err
		}
		cidString := link.Cid.String()
		if _, ok := visited[cidString]; ok {
			return nil
		}
		visited[cidString] = struct{}{}

		// We deliberately ignore the first returned value, which indicates whether
		// the block was actually verified or not. Unsigned blocks are still allowed,
		// but invalid embedded or sidecar signatures must reject the DAG.
		if _, err := coreblock.VerifyStoredBlockSignature(
			ctx,
			link.Cid,
			current,
			p.db.Multistore().Systemstore(),
			linkSys,
		); err != nil {
			return err
		}

		if _, err := linkSys.Store(linking.LinkContext{Ctx: ctx}, coreblock.GetLinkPrototype(), current.GenerateNode()); err != nil {
			return err
		}

		var encResults *encryption.Results
		if current.IsEncrypted() {
			results, err := p.kms.GetKeys(ctx, *current.Encryption)
			if err != nil {
				return err
			}
			encResults = results
		}

		for _, lnk := range current.AllLinks() {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			ctxWithTimeout, cancel := context.WithTimeout(ctx, p.syncBlockLinkTimeout)
			nd, err := linkSys.Load(linking.LinkContext{Ctx: ctxWithTimeout}, lnk, coreblock.BlockSchemaPrototype)
			cancel()

			if err != nil {
				return err
			}

			linkBlock, err := coreblock.GetFromNode(nd)
			if err != nil {
				return err
			}

			stack = append(stack, linkBlock)
		}

		if encResults != nil {
			for res := range encResults.Get() {
				if res.Error != nil {
					return res.Error
				}
			}
		}
		return nil
	}

	if err := processBlock(block); err != nil {
		return err
	}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if err := processBlock(current); err != nil {
			return err
		}
	}

	return nil
}
