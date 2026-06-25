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

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
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
