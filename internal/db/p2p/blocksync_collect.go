// Copyright 2026 Democratized Data Foundation
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

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

// collectBlocksForRequest gathers the blocks a peer needs in order to merge root.
//
// Starting from root it walks the DAG (composite history via Heads, field blocks via Links, plus
// each block's signature and encryption blocks) collecting every block reachable from root that is
// not behind one of haveHeads. When full is true, haveHeads is ignored and the whole reachable DAG
// is collected.
//
// dagBlocks are destined for the receiver's block store; encBlocks for its encryption store. A
// block that is not found locally is skipped — this node simply does not have that part of the
// history, and the receiver may obtain it from another peer.
func (p *P2P) collectBlocksForRequest(
	ctx context.Context,
	root cid.Cid,
	haveHeads []cid.Cid,
	full bool,
) (dagBlocks []blocks.Block, encBlocks []blocks.Block, err error) {
	bstore := p.db.Multistore().Blockstore()
	encstore := p.db.Multistore().Encstore()

	have := make(map[cid.Cid]struct{}, len(haveHeads))
	if !full {
		for _, h := range haveHeads {
			have[h] = struct{}{}
		}
	}

	visited := make(map[cid.Cid]struct{})
	queue := []cid.Cid{root}

	for len(queue) > 0 {
		if ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		current := queue[0]
		queue = queue[1:]

		if _, ok := visited[current]; ok {
			continue
		}
		if _, ok := have[current]; ok {
			continue
		}
		visited[current] = struct{}{}

		raw, err := bstore.Get(ctx, current)
		if err != nil {
			if ipld.IsNotFound(err) {
				continue
			}
			return nil, nil, err
		}
		dagBlocks = append(dagBlocks, raw)

		block, err := coreblock.GetFromBytes(raw.RawData())
		if err != nil {
			return nil, nil, NewErrDecodeLinkedBlock(err)
		}

		// Signature and encryption blocks are not part of AllLinks, so collect them explicitly.
		if block.Signature != nil {
			sigRaw, err := bstore.Get(ctx, block.Signature.Cid)
			if err != nil && !ipld.IsNotFound(err) {
				return nil, nil, err
			} else if err == nil {
				dagBlocks = append(dagBlocks, sigRaw)
			}
		}
		if block.Encryption != nil {
			encRaw, err := encstore.Get(ctx, block.Encryption.Cid)
			if err != nil && !ipld.IsNotFound(err) {
				return nil, nil, err
			} else if err == nil {
				encBlocks = append(encBlocks, encRaw)
			}
		}

		for _, lnk := range block.AllLinks() {
			if _, ok := visited[lnk.Cid]; ok {
				continue
			}
			if _, ok := have[lnk.Cid]; ok {
				continue
			}
			queue = append(queue, lnk.Cid)
		}
	}

	return dagBlocks, encBlocks, nil
}
