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
	"bytes"
	"context"
	"math"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	car "github.com/ipld/go-car/v2"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
)

// retentionGate refuses a pushed document before its blocks are fetched or written, using the height
// block its create head links. It refuses only on bytes that match the block's CID. Anything it
// cannot read is left to the check at merge.
//
// A gate serves one message and is not safe for concurrent use.
type retentionGate struct {
	p     *P2P
	field string
	floor int64
	// blocks caches height blocks by CID.
	blocks map[cid.Cid][]byte
	// fetched records the blocks already requested from the network.
	fetched map[cid.Cid]struct{}
}

// retentionGate returns the gate for collectionID, or nil if the rule does not judge it. A nil gate
// refuses nothing.
func (p *P2P) retentionGate(ctx context.Context, collectionID string) *retentionGate {
	if p.retentionFloor == nil {
		return nil
	}
	field, floor, ok := p.retentionFloor(ctx, collectionID)
	if !ok {
		return nil
	}
	return &retentionGate{
		p:       p,
		field:   field,
		floor:   floor,
		blocks:  make(map[cid.Cid][]byte),
		fetched: make(map[cid.Cid]struct{}),
	}
}

// heightLink returns the height block a create head links, or cid.Undef when the gate does not
// judge head.
func (g *retentionGate) heightLink(head *coreblock.Block) cid.Cid {
	if g == nil || head.Delta.GetPriority() != 1 {
		return cid.Undef
	}
	link, ok := head.GetLinkByName(g.field)
	if !ok {
		return cid.Undef
	}
	return link.Cid
}

// readCARs caches the blocks among links found in cars. The CAR reader verifies each block against
// its CID.
func (g *retentionGate) readCARs(cars [][]byte, links []cid.Cid) {
	if g == nil {
		return
	}
	wanted := make(map[cid.Cid]struct{}, len(links))
	for _, link := range links {
		if _, held := g.blocks[link]; link.Defined() && !held {
			wanted[link] = struct{}{}
		}
	}
	for _, data := range cars {
		if len(wanted) == 0 {
			return
		}
		if len(data) == 0 {
			continue
		}
		reader, err := car.NewBlockReader(bytes.NewReader(data))
		if err != nil {
			continue
		}
		for len(wanted) > 0 {
			block, err := reader.Next()
			if err != nil {
				break
			}
			if _, ok := wanted[block.Cid()]; ok {
				g.blocks[block.Cid()] = block.RawData()
				delete(wanted, block.Cid())
			}
		}
	}
}

// refusal returns the refusal for the document whose height block is link, or nil if it is kept or
// the block cannot be read. The block is looked up in the cache, then the local store, then, if
// fetch is set, the network.
func (g *retentionGate) refusal(ctx context.Context, link cid.Cid, fetch bool) error {
	if g == nil || !link.Defined() {
		return nil
	}
	raw, ok := g.heightBlock(ctx, link, fetch)
	if !ok {
		return nil
	}
	height, valid, ok := heightIn(raw, link, g.field)
	if !ok {
		return nil
	}
	return client.CheckRetention(height, valid, g.floor)
}

// heightBlock returns the raw block for link. The block service writes a fetched block to the local
// store.
func (g *retentionGate) heightBlock(ctx context.Context, link cid.Cid, fetch bool) ([]byte, bool) {
	if raw, ok := g.blocks[link]; ok {
		return raw, true
	}
	if block, err := g.p.db.Multistore().Blockstore().Get(ctx, link); err == nil {
		g.blocks[link] = block.RawData()
		return block.RawData(), true
	}
	if _, asked := g.fetched[link]; !fetch || asked {
		return nil, false
	}
	g.fetched[link] = struct{}{}

	g.p.statHeightFetches.Add(1)
	fetchCtx, cancel := context.WithTimeout(ctx, g.p.syncBlockLinkTimeout)
	defer cancel()
	raw, err := g.p.host.IPLDStore().Get(fetchCtx, link.KeyString())
	if err != nil {
		g.p.statHeightFetchMissed.Add(1)
		return nil, false
	}
	g.blocks[link] = raw
	return raw, true
}

// heightIn decodes the height a height block holds. ok is false when raw does not match link or is
// not a plain value of field; valid is false when the value is not a non-negative integer.
func heightIn(raw []byte, link cid.Cid, field string) (height int64, valid bool, ok bool) {
	sum, err := link.Prefix().Sum(raw)
	if err != nil || !sum.Equals(link) {
		return 0, false, false
	}
	block, err := coreblock.GetFromBytes(raw)
	if err != nil || block.Encryption != nil || block.Delta.LWWDelta == nil || block.Delta.GetFieldName() != field {
		return 0, false, false
	}
	var value any
	if err := cbor.Unmarshal(block.Delta.GetData(), &value); err != nil {
		return 0, false, false
	}
	// The decoder gives a non-negative integer as uint64, so any other type is not a height.
	v, isUint := value.(uint64)
	if !isUint || v > math.MaxInt64 {
		return 0, false, true
	}
	return int64(v), true, true
}

// retentionSkip returns the skip reason for a gate refusal made at stage. A document with no valid
// height has its own reason, whatever the stage.
func retentionSkip(err error, stage string) string {
	if errors.Is(err, client.ErrNoRetentionHeight) {
		return skipRetentionNoHeight
	}
	return stage
}
