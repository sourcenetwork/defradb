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
	"fmt"
	"io"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv/blockstore"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// peerIdentityFunc returns a function that resolves and caches the ACP identity of the given peer
// by fetching and verifying its identity token over the identity protocol.
func (p *P2P) peerIdentityFunc(ctx context.Context, pid string) func() immutable.Option[identity.Identity] {
	return func() immutable.Option[identity.Identity] {
		p.piMu.RLock()
		ident, ok := p.peerIdentities[pid]
		p.piMu.RUnlock()
		if ok {
			return immutable.Some(ident)
		}

		ctx, cancel := context.WithTimeout(ctx, networkRequestTimeout)
		defer cancel()
		resp, err := p.identityProtocol.GetIdentity(ctx, pid)
		if err != nil {
			log.ErrorE("Failed to get identity", err)
			return immutable.None[identity.Identity]()
		}
		ident, err = identity.FromToken(resp.IdentityToken)
		if err != nil {
			log.ErrorE("Failed to parse identity token", err)
			return immutable.None[identity.Identity]()
		}
		tokenIdent, ok := ident.(identity.TokenIdentity)
		if !ok {
			log.ErrorE("Identity is not of type TokenIdentity", nil,
				corelog.String("Actual", fmt.Sprintf("%T", ident)))
			return immutable.None[identity.Identity]()
		}
		if err := identity.VerifyAuthToken(tokenIdent, p.host.ID()); err != nil {
			log.ErrorE("Failed to verify auth token", err)
			return immutable.None[identity.Identity]()
		}
		p.piMu.Lock()
		p.peerIdentities[pid] = ident
		p.piMu.Unlock()
		return immutable.Some(ident)
	}
}

// peerHasDocAccess reports whether the given peer may read docID within the collection identified
// by collectionVersionID. Peers in the collection's replicator list are always granted access.
func (p *P2P) peerHasDocAccess(ctx context.Context, pid, docID, collectionVersionID string) (bool, error) {
	if !p.db.DocumentACP().HasValue() {
		return true, nil
	}

	ident, err := p.db.GetNodeIdentity(p.ctx)
	if err != nil {
		return false, err
	}
	// The collection lookup is a local operation on this node — authorise it as the node itself so
	// NAC sees a known identity rather than "anonymous".
	getColOpts := options.GetCollections().SetCollectionID(collectionVersionID)
	if ident.HasValue() {
		getColOpts = getColOpts.SetIdentity(identity.FromDID(ident.Value().DID))
	}
	cols, err := p.db.GetCollections(ctx, getColOpts)
	if err != nil {
		return false, err
	}
	if len(cols) == 0 {
		return false, client.ErrCollectionNotFound
	}

	// Replicators of this collection are always granted access.
	p.repMu.Lock()
	if peerList, ok := p.replicators[cols[0].CollectionID()]; ok {
		if _, exists := peerList[pid]; exists {
			p.repMu.Unlock()
			return true, nil
		}
	}
	p.repMu.Unlock()

	return acpDB.CheckDocAccessWithIdentityFunc(
		ctx,
		p.peerIdentityFunc(ctx, pid),
		p.db.NodeACP(),
		p.db.DocumentACP().Value(),
		cols[0],
		acpTypes.DocumentReadPerm,
		docID,
	)
}

// handleBlockSyncRequest serves an incoming block sync request: it enforces ACP for the requesting
// peer and, if allowed, returns the blocks needed to satisfy the request as a CAR.
//
// When the requester lacks access, or this node has none of the requested blocks, it returns a nil
// writeCAR (the requester sees an empty response and treats it as a no-op).
func (p *P2P) handleBlockSyncRequest(
	ctx context.Context,
	req protocol.BlockSyncRequest,
) ([][]byte, func(io.Writer) error, error) {
	root, err := cid.Cast(req.Root)
	if err != nil {
		return nil, nil, err
	}

	allowed, err := p.peerHasDocAccess(ctx, req.GetSenderID(), req.DocID, req.CollectionID)
	if err != nil {
		return nil, nil, err
	}
	if !allowed {
		return nil, nil, nil
	}

	haveHeads := make([]cid.Cid, 0, len(req.HaveHeads))
	for _, h := range req.HaveHeads {
		c, err := cid.Cast(h)
		if err != nil {
			return nil, nil, err
		}
		haveHeads = append(haveHeads, c)
	}

	dagBlocks, encBlocks, err := p.collectBlocksForRequest(ctx, root, haveHeads, req.Full)
	if err != nil {
		return nil, nil, err
	}
	if len(dagBlocks) == 0 {
		return nil, nil, nil
	}

	encCIDs := make([][]byte, len(encBlocks))
	for i, b := range encBlocks {
		encCIDs[i] = b.Cid().Bytes()
	}

	allBlocks := make([]blocks.Block, 0, len(dagBlocks)+len(encBlocks))
	allBlocks = append(allBlocks, dagBlocks...)
	allBlocks = append(allBlocks, encBlocks...)

	writeCARFn := func(w io.Writer) error {
		return writeCAR(ctx, w, []cid.Cid{root}, allBlocks)
	}
	return encCIDs, writeCARFn, nil
}

// pullAndIngest requests the blocks needed to merge root from fromPeer, ingests the returned CAR
// into the local stores, and prepares them for merge (signature verification + encryption keys).
//
// It is the bitswap-free replacement for syncDAG: the caller is responsible for calling db.Merge
// afterwards.
func (p *P2P) pullAndIngest(
	ctx context.Context,
	fromPeer string,
	docID string,
	collectionID string,
	root cid.Cid,
) error {
	haveHeads, err := p.localHeads(ctx, docID)
	if err != nil {
		return err
	}
	have := make([][]byte, len(haveHeads))
	for i, h := range haveHeads {
		have[i] = h.Bytes()
	}

	req := &protocol.BlockSyncRequest{
		DocID:        docID,
		CollectionID: collectionID,
		Root:         root.Bytes(),
		HaveHeads:    have,
		Full:         len(haveHeads) == 0,
	}

	blockDst := datastore.P2PBlockstoreFrom(p.db.Rootstore(), p.db.BlockStoreChunkSize())
	encDst := p.db.Multistore().Encstore()

	return p.blockSyncProtocol.RequestBlocks(ctx, fromPeer, req,
		func(encCIDs [][]byte, car io.Reader) error {
			encSet := make(map[cid.Cid]struct{}, len(encCIDs))
			for _, c := range encCIDs {
				parsed, err := cid.Cast(c)
				if err != nil {
					return err
				}
				encSet[parsed] = struct{}{}
			}
			if _, err := ingestCAR(ctx, blockDst, encDst, encSet, car); err != nil {
				return err
			}
			return p.prepareForMerge(ctx, root)
		},
	)
}

// localHeads returns the current composite head CIDs the local node has for the given document.
func (p *P2P) localHeads(ctx context.Context, docID string) ([]cid.Cid, error) {
	if docID == "" {
		return nil, nil
	}
	key := keys.HeadstoreDocKey{
		DocID:   docID,
		FieldID: core.COMPOSITE_NAMESPACE,
	}
	headset := coreblock.NewHeadSet(p.db.Multistore().Headstore(), key)
	cids, _, err := headset.List(ctx)
	return cids, err
}

// prepareForMerge walks the freshly-ingested DAG from root using a local link system, verifying
// block signatures and fetching any required encryption keys so that db.Merge can proceed.
func (p *P2P) prepareForMerge(ctx context.Context, root cid.Cid) error {
	localStore := blockstore.NewIPLDStore(p.db.Multistore().Blockstore())
	linkSys := makeLinkSystem(localStore)

	nd, err := linkSys.Load(linking.LinkContext{Ctx: ctx}, cidlink.Link{Cid: root}, coreblock.BlockSchemaPrototype)
	if err != nil {
		return NewErrLoadLinkedBlock(err)
	}
	rootBlock, err := coreblock.GetFromNode(nd)
	if err != nil {
		return NewErrDecodeLinkedBlock(err)
	}
	return p.loadBlockLinks(ctx, &linkSys, rootBlock)
}
