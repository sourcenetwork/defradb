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
	"github.com/sourcenetwork/corelog"
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

	p.statSyncDAGCalls.Add(1)

	// written counts the links this walk has loaded, which is how far it got rather than
	// how many blocks it added: a link already held loads without writing anything.
	var written int64

	// Store the block in the DAG store
	_, err := linkSystem.Store(linking.LinkContext{Ctx: sessionCtx}, coreblock.GetLinkPrototype(), block.GenerateNode())
	if err != nil {
		p.syncDAGFailure(reasonStoreRoot, written, err)
		return NewErrStoreBlockDAGSync(err)
	}
	written++

	reason, err := p.loadBlockLinks(sessionCtx, &linkSystem, block, &written)
	if err != nil {
		p.syncDAGFailure(reason, written, err)
		return err
	}
	return nil
}

// syncDAGFailure records an abandoned walk: how it failed and how far it had got. The
// first occurrence of each reason gets a log line; the rest are counted only.
func (p *P2P) syncDAGFailure(reason string, loaded int64, err error) {
	if p.syncDAGFailureReason.recordFirst(reason) {
		log.ErrorE("DAG sync abandoned", err,
			corelog.String("reason", reason),
			corelog.Int64("blocksLoaded", loaded))
	}
}

// loadBlockLinks traverses the DAG rooted at block and syncs all linked blocks.
// Uses an explicit stack to avoid goroutine stack overflow on deep DAGs (#2722).
//
// written is incremented for each block the walk pulls into the blockstore. On error the
// returned reason names the step that failed, for the caller's counters.
func (p *P2P) loadBlockLinks(
	ctx context.Context,
	linkSys *linking.LinkSystem,
	block *coreblock.Block,
	written *int64,
) (string, error) {
	bstore := datastore.BlockstoreFrom(p.db.Rootstore(), immutable.None[int]())
	stack := []*coreblock.Block{block}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		link, err := current.GenerateLink()
		if err != nil {
			return reasonBlockLink, NewErrGenerateBlockLink(err)
		}
		merged, err := bstore.IsMerged(ctx, link.Cid)
		if err != nil {
			return reasonIsMerged, NewErrCheckBlockMerged(err)
		}
		if merged {
			continue
		}

		// TODO: this part is not tested yet because there is not easy way of doing it at the moment.
		// https://github.com/sourcenetwork/defradb/issues/3525
		if current.Signature != nil {
			// we deliberately ignore the first returned value, which indicates whether the signature
			// the block was actually verified or not, because we don't handle it any different here.
			// But we want to keep the API of VerifyBlockSignature explicit about the results.
			_, err := coreblock.VerifyBlockSignature(current, linkSys)
			if err != nil {
				return reasonVerifySig, NewErrVerifyBlockSig(err)
			}
		}

		var encResults *encryption.Results
		if current.IsEncrypted() {
			results, err := p.kms.GetKeys(ctx, *current.Encryption)
			if err != nil {
				return reasonEncKeys, NewErrGetEncKeysForBlock(err)
			}
			encResults = results
		}

		for _, lnk := range current.AllLinks() {
			if ctx.Err() != nil {
				return reasonContext, ctx.Err()
			}

			// Skip fetch if the linked block is already merged locally — avoids a BitSwap
			// round-trip for historical blocks that may have been pruned on the sender.
			linkedMerged, err := bstore.IsMerged(ctx, lnk.Cid)
			if err != nil {
				return reasonIsMerged, NewErrCheckBlockMerged(err)
			}
			if linkedMerged {
				continue
			}

			ctxWithTimeout, cancel := context.WithTimeout(ctx, p.syncBlockLinkTimeout)
			nd, err := linkSys.Load(linking.LinkContext{Ctx: ctxWithTimeout}, lnk, coreblock.BlockSchemaPrototype)
			cancel()

			if err != nil {
				return reasonLoadLink, NewErrLoadLinkedBlock(err)
			}
			// The link system writes what it fetches, so a successful load is one more
			// block in the store.
			*written++

			linkBlock, err := coreblock.GetFromNode(nd)
			if err != nil {
				return reasonDecodeLink, NewErrDecodeLinkedBlock(err)
			}

			stack = append(stack, linkBlock)
		}

		if encResults != nil {
			for res := range encResults.Get() {
				if res.Error != nil {
					return reasonEncKeys, NewErrRetrieveEncKey(res.Error)
				}
			}
		}
	}

	return "", nil
}
