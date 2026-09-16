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
	"strings"
	"sync"
	"time"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	car "github.com/ipld/go-car/v2"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/errors"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
)

// An update is announced by its head CID and root block, never its CAR. A receiver that finds
// it needs the head asks for the CAR, and the peer asked builds it then. A CAR is therefore sent
// only to a peer that asked for it, rather than to every subscriber whether or not it already
// holds the document.
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

	// carPeerBackoffBase is the first window after a peer's CAR request fails. Heads arrive far
	// faster than a window measured in seconds: at ~8 heads/second a 30s pass-over covers about
	// 240 of them, so one failure must cost only the next few.
	carPeerBackoffBase = 250 * time.Millisecond

	// carPeerBackoffMax caps the window, which doubles with each consecutive failure.
	carPeerBackoffMax = 30 * time.Second

	// carPeerBackoffHardFailures is how many consecutive failures make a peer known-bad rather
	// than merely slow. Below it a backed-off peer is still asked, after every other source,
	// because the creator is usually the only peer holding the head. At or above it the peer is
	// passed over, so a receiver with no route to it stops spending a request timeout per
	// message.
	carPeerBackoffHardFailures = 3
)

// Which peer a CAR was asked of. The creator published the update, and the relay is the peer
// that delivered it, when that is someone else.
const (
	carSourceCreator = "creator"
	carSourceRelay   = "relay"
)

// What became of one head asked of one peer, recorded as source and outcome together, for
// example "creatorNotServed". Both halves are fixed sets.
const (
	fetchFetched       = "Fetched"
	fetchNotServed     = "NotServed"
	fetchRootMismatch  = "RootMismatch"
	fetchRequestFailed = "RequestFailed"
	fetchTimeout       = "Timeout"
	fetchDialFailed    = "DialFailed"
	fetchCanceled      = "Canceled"
	fetchBadReply      = "BadReply"
	fetchBackoff       = "Backoff"
)

// carSourceRetrySuffix marks a source asked only after every other source came up short, because
// it is inside a backoff window. Its outcomes count under their own labels, so heads recovered
// from a backed-off creator are visible as creatorRetryFetched rather than hidden.
const carSourceRetrySuffix = "Retry"

// What the serving side did with one requested head.
const (
	serveServed      = "served"
	serveDeferred    = "deferred"
	serveBadCID      = "badCID"
	serveNotHeld     = "notHeld"
	serveReadFailed  = "readFailed"
	serveNoAccess    = "noAccess"
	serveNotDocument = "notDocument"
	serveTooLarge    = "tooLarge"
	serveBuildFailed = "buildFailed"
	serveContext     = "context"
)

var (
	errCARHeadNotHeld     = errors.New("requested CAR head is not held")
	errCARHeadNotDocument = errors.New("requested CAR head is not a document block")
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
	bstore := datastore.BlockstoreFrom(p.db.Rootstore(), immutable.None[int]())
	budget := maxCARReplyBytes
	for i, raw := range rawCIDs {
		if ctx.Err() != nil {
			reply.CARs = reply.CARs[:i]
			p.carServeOutcome.recordN(serveContext, int64(len(rawCIDs)-i))
			break
		}
		head, err := cid.Cast(raw)
		if err != nil {
			p.carServeOutcome.record(serveBadCID)
			continue
		}
		// Checked ahead of access, which also fails for a head not held and would hide the
		// difference between a peer asking the wrong node and a peer refused.
		held, err := bstore.Has(ctx, head)
		if err != nil {
			p.carServeOutcome.record(serveReadFailed)
			continue
		}
		if !held {
			p.carServeOutcome.record(serveNotHeld)
			continue
		}
		// The same gate bitswap applies before sending a block, so a CAR hands over nothing the
		// peer could not already fetch block by block.
		if !p.hasAccess(ctx, peerID, head) {
			p.carServeOutcome.record(serveNoAccess)
			continue
		}
		data, err := p.carForHead(ctx, head)
		if err != nil {
			p.carServeOutcome.record(carServeFailure(err))
			continue
		}
		if len(data) > maxCARReplyBytes {
			p.carServeOutcome.record(serveTooLarge)
			continue
		}
		if len(data) > budget {
			reply.CARs = reply.CARs[:i]
			p.carServeOutcome.recordN(serveDeferred, int64(len(rawCIDs)-i))
			break
		}
		budget -= len(data)
		reply.CARs[i] = data
		p.carServeOutcome.record(serveServed)
	}
	return reply
}

// carServeFailure names why carForHead returned no CAR.
func carServeFailure(err error) string {
	switch {
	case errors.Is(err, errCARHeadNotHeld):
		return serveNotHeld
	case errors.Is(err, errCARHeadNotDocument):
		return serveNotDocument
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return serveContext
	default:
		return serveBuildFailed
	}
}

// carForHead returns the CAR rooted at head, built from the local store unless a recent request
// already built it. The head must decode as a document block, which keeps the served set to what
// an update could have announced.
func (p *P2P) carForHead(ctx context.Context, head cid.Cid) ([]byte, error) {
	data, shared, err := p.carCache.getOrBuild(ctx, head.KeyString(), func() ([]byte, error) {
		bstore := datastore.BlockstoreFrom(p.db.Rootstore(), immutable.None[int]())
		raw, err := bstore.Get(ctx, head)
		if err != nil {
			if ipld.IsNotFound(err) {
				return nil, errCARHeadNotHeld
			}
			// Counted as buildFailed by the serve outcomes, which alone say nothing about the
			// cause. Every other build failure records a reason; this one did not.
			p.carFailure(reasonBlockRead, err)
			return nil, err
		}
		block, err := coreblock.GetFromBytes(raw.RawData())
		if err != nil {
			return nil, errCARHeadNotDocument
		}
		return p.generateCAR(ctx, block)
	})
	if shared {
		p.statCARCacheHits.Add(1)
	}
	return data, err
}

// carSource is one peer to ask for CARs, labelled for the fetch outcome counters.
type carSource struct {
	peerID string
	label  string
}

// carSources orders the peers to ask for an update's CARs. The creator comes first: relayed
// updates are never republished, so the creator is the one peer sure to hold the head when the
// update is announced. The relay comes second, and only when it is someone else. Pubsub hands on
// a message as soon as it arrives, before the relaying node has stored the head itself, so a relay
// asked at once rarely has it; by the time the creator has failed, it may.
//
// The creator is named inside the update rather than authenticated by pubsub. A peer naming
// another node as creator can only send receivers to it for heads they need, and every CAR is
// checked against the head it was asked for.
func carSources(creator, relay, self string) []carSource {
	var sources []carSource
	if creator != "" && creator != self {
		sources = append(sources, carSource{peerID: creator, label: carSourceCreator})
	}
	if relay != "" && relay != self && relay != creator {
		sources = append(sources, carSource{peerID: relay, label: carSourceRelay})
	}
	return sources
}

// fetchCARs asks for the CARs rooted at heads, from the update's creator and then from the peer
// that relayed it, each asked only for the heads still without one. The result is parallel to
// heads, and an entry is nil when no peer served a CAR rooted at that head; the caller walks the
// DAG for those. carFetchMissed therefore counts exactly the heads that fall back to the walk.
func (p *P2P) fetchCARs(ctx context.Context, creator, relay string, heads []cid.Cid) [][]byte {
	cars := make([][]byte, len(heads))
	if len(heads) == 0 || p.carProtocol == nil {
		return cars
	}

	pending := make([]int, len(heads))
	for i := range pending {
		pending[i] = i
	}
	ordered, passedOver := p.orderCARSources(carSources(creator, relay, p.host.ID()))
	for _, source := range ordered {
		if len(pending) == 0 {
			break
		}
		pending = p.fetchCARsFrom(ctx, source, heads, cars, pending)
	}
	// Heads still without a CAR that a passed-over peer was never asked for are what its backoff
	// cost this message.
	for _, source := range passedOver {
		if len(pending) == 0 {
			break
		}
		p.carFetchOutcome.recordN(source.label+fetchBackoff, int64(len(pending)))
	}

	p.statCARFetched.Add(int64(len(heads) - len(pending)))
	p.statCARFetchMissed.Add(int64(len(pending)))
	return cars
}

// orderCARSources splits sources into the order to ask them in and those to pass over. A peer
// inside a backoff window is moved behind every ready peer rather than dropped: the relay has
// usually not stored the head yet, so excluding a backed-off creator sends the head to a DAG walk
// when the one peer holding it was merely slow a moment ago. Only a peer that has failed
// carPeerBackoffHardFailures times in a row is dropped.
func (p *P2P) orderCARSources(sources []carSource) (ordered, passedOver []carSource) {
	var retry []carSource
	for _, source := range sources {
		switch p.carPeerAvailability(source.peerID) {
		case carPeerReady:
			ordered = append(ordered, source)
		case carPeerDeprioritised:
			source.label += carSourceRetrySuffix
			retry = append(retry, source)
		default:
			passedOver = append(passedOver, source)
		}
	}
	return append(ordered, retry...), passedOver
}

// fetchCARsFrom asks source for the heads at the pending indexes, fills what it serves into cars,
// and returns the indexes still without a CAR. It normally takes one round trip, and another each
// time a reply stops short. A request that fails through the peer's own fault backs it off; the
// caller has already decided whether a backed-off peer is asked at all.
func (p *P2P) fetchCARsFrom(
	ctx context.Context,
	source carSource,
	heads []cid.Cid,
	cars [][]byte,
	pending []int,
) []int {
	var unserved []int
	// Each reply answers at least one head or ends the loop, so it runs at most len(pending) times.
	for answered := 0; answered < len(pending); {
		ask := pending[answered:]
		got, outcome := p.requestCARs(ctx, source.peerID, heads, ask)
		if outcome != "" {
			p.carFetchOutcome.recordN(source.label+outcome, int64(len(ask)))
			if carFetchFailedPeer(outcome) {
				p.backOffCARPeer(source.peerID)
			}
			return append(unserved, ask...)
		}
		// The peer answered, so whatever it failed on before has passed.
		p.clearCARPeerBackoff(source.peerID)
		for i, data := range got {
			idx := ask[i]
			switch {
			case len(data) == 0:
				p.carFetchOutcome.record(source.label + fetchNotServed)
				unserved = append(unserved, idx)
			case !carRootIs(data, heads[idx]):
				p.carFetchOutcome.record(source.label + fetchRootMismatch)
				unserved = append(unserved, idx)
			default:
				p.carFetchOutcome.record(source.label + fetchFetched)
				cars[idx] = data
			}
		}
		answered += len(got)
	}
	return unserved
}

// requestCARs sends one request for the heads at the given indexes and returns the reply's
// entries, which cover a prefix of them. It returns a fetch outcome instead when the request
// failed or the reply answered nothing usable, either of which ends the asking of this peer.
func (p *P2P) requestCARs(ctx context.Context, peerID string, heads []cid.Cid, ask []int) ([][]byte, string) {
	req := protocol.CARRequest{CIDs: make([][]byte, len(ask))}
	for i, idx := range ask {
		req.CIDs[i] = heads[idx].Bytes()
	}

	parent := ctx
	ctx, cancel := context.WithTimeout(ctx, networkRequestTimeout)
	defer cancel()
	reply, err := p.carProtocol.SendRequest(ctx, req, peerID)
	if err != nil {
		outcome := carRequestFailure(parent, err)
		// The counters say how often a request failed but not which peer or why, and a timeout
		// under load needs a different fix from a peer with no route. Logged once per distinct
		// outcome, so a persistent failure is a count rather than a line per call.
		if p.carFetchFailureReason.recordFirst(outcome) {
			log.ErrorE("CAR request failed", err,
				corelog.String("outcome", outcome),
				corelog.String("peer", peerID),
				corelog.Int("heads", len(ask)))
		}
		return nil, outcome
	}
	if reply.GetErrMessage() != "" || len(reply.CARs) == 0 || len(reply.CARs) > len(ask) {
		return nil, fetchBadReply
	}
	return reply.CARs, ""
}

// carRequestFailure names why a request did not come back. The parent context ending is the local
// node shutting down or giving up on the message, which is not the peer's doing and must not be
// reported or counted as the peer failing.
func carRequestFailure(parent context.Context, err error) string {
	switch {
	case parent.Err() != nil:
		return fetchCanceled
	case errors.Is(err, context.DeadlineExceeded):
		return fetchTimeout
	// libp2p wraps dial failures rather than exporting a sentinel for them.
	case strings.Contains(err.Error(), "dial"):
		return fetchDialFailed
	default:
		return fetchRequestFailed
	}
}

// carFetchFailedPeer reports whether an outcome should count against the peer. A reply that came
// back malformed, and a request this node abandoned, are not the peer failing to answer.
func carFetchFailedPeer(outcome string) bool {
	switch outcome {
	case fetchRequestFailed, fetchTimeout, fetchDialFailed:
		return true
	default:
		return false
	}
}

// carPeerBackoffs records peers whose CAR request failed: until when they are held back, and how
// many times in a row they have failed.
type carPeerBackoffs struct {
	mu    sync.Mutex
	peers map[string]*carPeerBackoff
}

// carPeerBackoff is one peer's consecutive failures and the window they bought.
type carPeerBackoff struct {
	until    time.Time
	failures int
}

// carPeerAvailability is how a peer should be treated for one fetch.
type carPeerAvailability int

const (
	// carPeerReady is asked in the normal order.
	carPeerReady carPeerAvailability = iota
	// carPeerDeprioritised recently failed, so is asked only after every ready source.
	carPeerDeprioritised
	// carPeerPassedOver has failed repeatedly, so is not asked until its window expires.
	carPeerPassedOver
)

// carPeerAvailability reports how peerID should be treated now, and forgets a peer whose window
// has expired.
func (p *P2P) carPeerAvailability(peerID string) carPeerAvailability {
	p.carBackoff.mu.Lock()
	defer p.carBackoff.mu.Unlock()
	entry, ok := p.carBackoff.peers[peerID]
	if !ok {
		return carPeerReady
	}
	if !time.Now().Before(entry.until) {
		delete(p.carBackoff.peers, peerID)
		return carPeerReady
	}
	if entry.failures >= carPeerBackoffHardFailures {
		return carPeerPassedOver
	}
	return carPeerDeprioritised
}

// backOffCARPeer records a failed request against peerID. The window doubles with each
// consecutive failure, so an isolated failure costs the next few heads and a peer that is
// genuinely gone reaches carPeerBackoffMax quickly.
func (p *P2P) backOffCARPeer(peerID string) {
	p.carBackoff.mu.Lock()
	defer p.carBackoff.mu.Unlock()
	if p.carBackoff.peers == nil {
		p.carBackoff.peers = make(map[string]*carPeerBackoff)
	}
	entry, ok := p.carBackoff.peers[peerID]
	if !ok {
		entry = &carPeerBackoff{}
		p.carBackoff.peers[peerID] = entry
	}
	entry.failures++
	window := carPeerBackoffMax
	// Bounded so the shift cannot overflow; base<<7 is already past the cap.
	if shift := entry.failures - 1; shift < 8 {
		if grown := carPeerBackoffBase << shift; grown < carPeerBackoffMax {
			window = grown
		}
	}
	entry.until = time.Now().Add(window)
}

// clearCARPeerBackoff forgets peerID's failures once it answers, so a peer that is serving again
// is not left deprioritised by a failure it has recovered from.
func (p *P2P) clearCARPeerBackoff(peerID string) {
	p.carBackoff.mu.Lock()
	defer p.carBackoff.mu.Unlock()
	delete(p.carBackoff.peers, peerID)
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
