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

	"github.com/ipfs/go-cid"
	car "github.com/ipld/go-car/v2"

	"github.com/sourcenetwork/immutable"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
)

// An update is announced by its head CID and root block, never its CAR. A receiver that finds
// it needs the head asks the sender for the CAR, and the sender builds it then. A CAR is
// therefore sent only to a peer that asked for it, rather than to every subscriber whether or
// not it already holds the document.
const (
	// maxCARRequestCIDs caps the heads one reply attempts. It matches the largest batch a
	// receiver accepts, so one request normally covers a whole inbound message whatever batch
	// size the sender used. It is not a cap on what a receiver can fetch: heads past it are
	// left out of the reply and asked for again.
	maxCARRequestCIDs = maxInboundDocuments

	// maxCARReplyBytes caps the CAR bytes in one reply, under the direct-stream message cap so
	// the envelope still fits. The reply stops at the first CAR that would pass it, and the
	// requester asks again for the rest. A single CAR larger than the whole cap is never sent,
	// and the requester walks the DAG for that head.
	maxCARReplyBytes = 8 << 20
)

// carCommProcessor serves CAR requests from peers.
type carCommProcessor struct {
	p2p *P2P
}

func (proc *carCommProcessor) ProcessRequest(
	ctx context.Context,
	req protocol.CARRequest,
) (protocol.CARReply, error) {
	return proc.p2p.serveCARs(ctx, req.SenderID, req.CIDs), nil
}

// serveCARs builds the CAR for each requested head the peer may read, in request order. A head
// that is malformed, not held, not readable by the peer, or larger than a whole reply gets an
// empty entry rather than failing the reply, so one bad head does not cost the peer the others.
//
// The reply ends early at the count cap, or at the first CAR that would pass the byte budget,
// and the peer asks again for the heads past its end. A CAR built and left out that way stays
// in the cache, so the next request does not build it twice.
func (p *P2P) serveCARs(ctx context.Context, peerID string, rawCIDs [][]byte) protocol.CARReply {
	if len(rawCIDs) > maxCARRequestCIDs {
		rawCIDs = rawCIDs[:maxCARRequestCIDs]
	}
	reply := protocol.CARReply{CARs: make([][]byte, len(rawCIDs))}
	budget := maxCARReplyBytes
	for i, raw := range rawCIDs {
		if ctx.Err() != nil {
			reply.CARs = reply.CARs[:i]
			break
		}
		head, err := cid.Cast(raw)
		if err != nil {
			continue
		}
		// The same gate bitswap applies before sending a block, so a CAR hands over nothing the
		// peer could not already fetch block by block.
		if !p.hasAccess(ctx, peerID, head) {
			continue
		}
		data, err := p.carForHead(ctx, head)
		if err != nil || len(data) > maxCARReplyBytes {
			continue
		}
		if len(data) > budget {
			reply.CARs = reply.CARs[:i]
			break
		}
		budget -= len(data)
		reply.CARs[i] = data
	}
	return reply
}

// carForHead returns the CAR rooted at head, built from the local store unless a recent request
// already built it. The head must decode as a document block, which keeps the served set to what
// an update could have announced.
func (p *P2P) carForHead(ctx context.Context, head cid.Cid) ([]byte, error) {
	data, shared, err := p.carCache.getOrBuild(ctx, head.KeyString(), func() ([]byte, error) {
		bstore := datastore.BlockstoreFrom(p.db.Rootstore(), immutable.None[int]())
		raw, err := bstore.Get(ctx, head)
		if err != nil {
			return nil, err
		}
		block, err := coreblock.GetFromBytes(raw.RawData())
		if err != nil {
			return nil, err
		}
		return p.generateCAR(ctx, block)
	})
	if shared {
		p.statCARCacheHits.Add(1)
	}
	return data, err
}

// fetchCARs asks peerID for the CARs rooted at heads. It normally takes one round trip, and
// another each time a reply stops short, until every head has been answered. The result is
// parallel to heads, and an entry is nil when the peer did not serve it or served a CAR for
// another root; the caller walks the DAG for those. A peer that predates the exchange fails the
// request, which leaves every entry nil.
func (p *P2P) fetchCARs(ctx context.Context, peerID string, heads []cid.Cid) [][]byte {
	cars := make([][]byte, len(heads))
	if len(heads) == 0 || peerID == "" || p.carProtocol == nil || peerID == p.host.ID() {
		return cars
	}

	// Each reply answers at least one head or ends the loop, so it runs at most len(heads) times.
	for answered := 0; answered < len(heads); {
		n, ok := p.fetchCARsOnce(ctx, peerID, heads[answered:], cars[answered:])
		if !ok {
			p.statCARFetchMissed.Add(int64(len(heads) - answered))
			break
		}
		answered += n
	}
	return cars
}

// fetchCARsOnce sends one request for heads and fills the verified CARs of the prefix it answered
// into cars, returning how many heads that prefix covers. It reports false for a failed request
// or a reply that answered nothing, either of which ends the fetch.
func (p *P2P) fetchCARsOnce(ctx context.Context, peerID string, heads []cid.Cid, cars [][]byte) (int, bool) {
	req := protocol.CARRequest{CIDs: make([][]byte, len(heads))}
	for i, head := range heads {
		req.CIDs[i] = head.Bytes()
	}

	ctx, cancel := context.WithTimeout(ctx, networkRequestTimeout)
	defer cancel()
	reply, err := p.carProtocol.SendRequest(ctx, req, peerID)
	if err != nil || reply.GetErrMessage() != "" || len(reply.CARs) == 0 || len(reply.CARs) > len(heads) {
		return 0, false
	}

	for i, data := range reply.CARs {
		if len(data) == 0 || !carRootIs(data, heads[i]) {
			p.statCARFetchMissed.Add(1)
			continue
		}
		cars[i] = data
		p.statCARFetched.Add(1)
	}
	return len(reply.CARs), true
}

// carRootIs reports whether data is a CAR whose sole root is head. A fetched CAR is imported in
// place of walking the DAG from a root block that has already passed its checks, so it must be
// rooted at that same block.
func carRootIs(data []byte, head cid.Cid) bool {
	reader, err := car.NewBlockReader(bytes.NewReader(data))
	if err != nil {
		return false
	}
	return len(reader.Roots) == 1 && reader.Roots[0].Equals(head)
}
