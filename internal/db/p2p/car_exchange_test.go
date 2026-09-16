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
	"errors"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/internal/db/p2p/protocol"
)

// servingDB adds the one method the serving side reaches beyond the store: without document
// ACP every peer may read every block.
type servingDB struct {
	rootstoreDB
}

func (servingDB) DocumentACP() immutable.Option[dac.DocumentACP] {
	return immutable.None[dac.DocumentACP]()
}

// serveEveryBlock swaps p's DB for one without document ACP, so the serving side reads every
// block. carFixture always builds p with a rootstoreDB.
func serveEveryBlock(t *testing.T, p *P2P) {
	t.Helper()
	store, ok := p.db.(rootstoreDB)
	require.True(t, ok, "expected carFixture to provide a rootstoreDB, got %T", p.db)
	p.db = servingDB{rootstoreDB: store}
}

// carCall is one CAR request a fakeCARChannel received: the peer asked and how many heads.
type carCall struct {
	peer  string
	heads int
}

// fakeCARChannel answers each peer's CAR requests with that peer's replies in turn. A peer with
// no replies left, or none at all, fails the request as an unreachable peer would.
type fakeCARChannel struct {
	replies   map[string][]protocol.CARReply
	requested []carCall
}

func (f *fakeCARChannel) SendRequest(
	_ context.Context,
	req protocol.CARRequest,
	peerID string,
) (protocol.CARReply, error) {
	f.requested = append(f.requested, carCall{peer: peerID, heads: len(req.CIDs)})
	queue := f.replies[peerID]
	if len(queue) == 0 {
		return protocol.CARReply{}, errors.New("peer unreachable")
	}
	f.replies[peerID] = queue[1:]
	return queue[0], nil
}

// fetchFixture returns a P2P, whose own ID is "peerID", that asks channel for CARs.
func fetchFixture(channel *fakeCARChannel) *P2P {
	return withReasonMaps(&P2P{host: &SimpleMockHost{}, carProtocol: channel})
}

// A held head is served as a CAR rooted at it. A head this node does not hold, or bytes that
// are not a CID, get an empty entry without costing the requester the rest of the reply, and
// each is counted under its own reason.
func TestServeCARs_ServesHeldHeadsAndLeavesTheRestEmpty(t *testing.T) {
	child := undecodableBlock(t, "field block")
	_, root := compositeLinking(t, child.Cid())
	p := carFixture(t, root, child)
	serveEveryBlock(t, p)

	absent := compositeBlock(t, 7)
	reply := p.serveCARs(context.Background(), "peer", [][]byte{
		root.Cid().Bytes(),
		absent.Cid().Bytes(),
		[]byte("not a cid"),
	})

	require.Len(t, reply.CARs, 3)
	require.True(t, carRootIs(reply.CARs[0], root.Cid()))
	require.Empty(t, reply.CARs[1])
	require.Empty(t, reply.CARs[2])
	require.Equal(t, int64(1), p.statCARBuilt.Load())
	require.Equal(t,
		map[string]int64{serveServed: 1, serveNotHeld: 1, serveBadCID: 1},
		reasonMap(p.carServeOutcome.drain()))
}

// A held block that is not a document block is not served.
func TestServeCARs_DoesNotServeANonDocumentHead(t *testing.T) {
	notDocument := undecodableBlock(t, "not a block")
	p := carFixture(t, notDocument)
	serveEveryBlock(t, p)

	reply := p.serveCARs(context.Background(), "peer", [][]byte{notDocument.Cid().Bytes()})

	require.Equal(t, [][]byte{nil}, reply.CARs)
	require.Equal(t, map[string]int64{serveNotDocument: 1}, reasonMap(p.carServeOutcome.drain()))
}

// A request for more heads than a batch holds is answered only up to the cap.
func TestServeCARs_AnswersAtMostTheCap(t *testing.T) {
	p := carFixture(t)
	serveEveryBlock(t, p)

	absent := compositeBlock(t, 1)
	req := make([][]byte, maxCARRequestCIDs+5)
	for i := range req {
		req[i] = absent.Cid().Bytes()
	}

	reply := p.serveCARs(context.Background(), "peer", req)
	require.Len(t, reply.CARs, maxCARRequestCIDs)
}

// A batch larger than the sender's batch size is served in one reply, not cut to it.
func TestServeCARs_AnswersABatchLargerThanTheSenderBatchSize(t *testing.T) {
	child := undecodableBlock(t, "field block")
	_, root := compositeLinking(t, child.Cid())
	p := carFixture(t, root, child)
	serveEveryBlock(t, p)
	p.carCache = newCARCache(carCacheMaxBytes)

	req := make([][]byte, 4*batchMaxDocs)
	for i := range req {
		req[i] = root.Cid().Bytes()
	}

	reply := p.serveCARs(context.Background(), "peer", req)
	require.Len(t, reply.CARs, len(req))
	require.True(t, carRootIs(reply.CARs[len(req)-1], root.Cid()))
}

// A served CAR is kept only when rooted at the head it was asked for. One rooted elsewhere, or
// not served, leaves the entry nil so the caller walks the DAG for it.
func TestFetchCARs_KeepsOnlyCARsRootedAtTheirHead(t *testing.T) {
	_, first := compositeLinking(t)
	second := compositeBlock(t, 2)
	third := compositeBlock(t, 3)

	channel := &fakeCARChannel{replies: map[string][]protocol.CARReply{"creator": {{CARs: [][]byte{
		carWith(t, first.Cid(), first),
		carWith(t, third.Cid(), third), // rooted at the wrong head
		nil,
	}}}}}
	p := fetchFixture(channel)

	cars := p.fetchCARs(context.Background(), "creator", "creator",
		[]cid.Cid{first.Cid(), second.Cid(), third.Cid()})

	require.Len(t, cars, 3)
	require.NotNil(t, cars[0])
	require.Nil(t, cars[1])
	require.Nil(t, cars[2])
	require.Equal(t, int64(1), p.statCARFetched.Load())
	require.Equal(t, int64(2), p.statCARFetchMissed.Load())
	require.Equal(t, map[string]int64{
		"creatorFetched":      1,
		"creatorRootMismatch": 1,
		"creatorNotServed":    1,
	}, reasonMap(p.carFetchOutcome.drain()))
}

// The creator is asked first, and when it serves everything the relay is never asked.
func TestFetchCARs_AsksTheCreatorBeforeTheRelay(t *testing.T) {
	head := compositeBlock(t, 1)
	channel := &fakeCARChannel{replies: map[string][]protocol.CARReply{
		"creator": {{CARs: [][]byte{carWith(t, head.Cid(), head)}}},
		"relay":   {{CARs: [][]byte{carWith(t, head.Cid(), head)}}},
	}}
	p := fetchFixture(channel)

	cars := p.fetchCARs(context.Background(), "creator", "relay", []cid.Cid{head.Cid()})

	require.NotNil(t, cars[0])
	require.Equal(t, []carCall{{peer: "creator", heads: 1}}, channel.requested)
}

// The relay is asked only for the heads the creator did not serve.
func TestFetchCARs_AsksTheRelayOnlyForHeadsTheCreatorDidNotServe(t *testing.T) {
	first := compositeBlock(t, 1)
	second := compositeBlock(t, 2)
	channel := &fakeCARChannel{replies: map[string][]protocol.CARReply{
		"creator": {{CARs: [][]byte{carWith(t, first.Cid(), first), nil}}},
		"relay":   {{CARs: [][]byte{carWith(t, second.Cid(), second)}}},
	}}
	p := fetchFixture(channel)

	cars := p.fetchCARs(context.Background(), "creator", "relay", []cid.Cid{first.Cid(), second.Cid()})

	require.True(t, carRootIs(cars[0], first.Cid()))
	require.True(t, carRootIs(cars[1], second.Cid()))
	require.Equal(t, []carCall{{peer: "creator", heads: 2}, {peer: "relay", heads: 1}}, channel.requested)
	require.Equal(t, int64(2), p.statCARFetched.Load())
	require.Zero(t, p.statCARFetchMissed.Load())
	require.Equal(t, map[string]int64{
		"creatorFetched":   1,
		"creatorNotServed": 1,
		"relayFetched":     1,
	}, reasonMap(p.carFetchOutcome.drain()))
}

// A creator whose request failed drops behind the relay on the next fetch, so a peer known to be
// serving is asked first rather than a worker being held for another request timeout.
func TestFetchCARs_PrefersAReadyRelayOverARecentlyFailedCreator(t *testing.T) {
	head := compositeBlock(t, 1)
	channel := &fakeCARChannel{replies: map[string][]protocol.CARReply{
		"relay": {
			{CARs: [][]byte{carWith(t, head.Cid(), head)}},
			{CARs: [][]byte{carWith(t, head.Cid(), head)}},
		},
	}}
	p := fetchFixture(channel)

	p.fetchCARs(context.Background(), "creator", "relay", []cid.Cid{head.Cid()})
	cars := p.fetchCARs(context.Background(), "creator", "relay", []cid.Cid{head.Cid()})

	require.NotNil(t, cars[0])
	// The failed creator drops behind the relay, and the relay serving the head means it is
	// never reached.
	require.Equal(t, []carCall{
		{peer: "creator", heads: 1},
		{peer: "relay", heads: 1},
		{peer: "relay", heads: 1},
	}, channel.requested)
	require.Equal(t, map[string]int64{
		"creatorRequestFailed": 1,
		"relayFetched":         2,
	}, reasonMap(p.carFetchOutcome.drain()))
}

// The headline fix: a creator inside a backoff window is still asked once the relay has come up
// short. The relay usually has not stored the head yet, so passing the creator over entirely sent
// the head to a DAG walk when the one peer holding it was merely slow a moment ago.
func TestFetchCARs_AsksABackedOffCreatorTheRelayCannotServe(t *testing.T) {
	head := compositeBlock(t, 1)
	channel := &fakeCARChannel{replies: map[string][]protocol.CARReply{
		// The relay is reachable throughout but holds nothing, as a relay asked at once usually
		// does not: pubsub hands a message on before the relaying node has stored the head.
		"relay": {{CARs: [][]byte{nil}}, {CARs: [][]byte{nil}}},
	}}
	p := fetchFixture(channel)

	// The creator has no reply queued, so this request fails and backs it off.
	p.fetchCARs(context.Background(), "creator", "relay", []cid.Cid{head.Cid()})
	missedBefore := p.statCARFetchMissed.Load()

	channel.replies["creator"] = []protocol.CARReply{{CARs: [][]byte{carWith(t, head.Cid(), head)}}}
	cars := p.fetchCARs(context.Background(), "creator", "relay", []cid.Cid{head.Cid()})

	require.NotNil(t, cars[0], "a backed-off creator must still be asked when nobody else can serve")
	require.Equal(t, []carCall{
		{peer: "creator", heads: 1},
		{peer: "relay", heads: 1},
		{peer: "relay", heads: 1},
		{peer: "creator", heads: 1},
	}, channel.requested)
	require.Equal(t, map[string]int64{
		"creatorRequestFailed": 1,
		"relayNotServed":       2,
		"creatorRetryFetched":  1,
	}, reasonMap(p.carFetchOutcome.drain()))
	require.Equal(t, missedBefore, p.statCARFetchMissed.Load(), "the head must not fall back to a walk")
}

// A peer that keeps failing is eventually passed over, so a receiver with no route to it stops
// spending a request timeout per message.
func TestFetchCARs_PassesOverAPeerThatKeepsFailing(t *testing.T) {
	head := compositeBlock(t, 1)
	channel := &fakeCARChannel{}
	p := fetchFixture(channel)

	for range carPeerBackoffHardFailures {
		p.fetchCARs(context.Background(), "creator", "", []cid.Cid{head.Cid()})
	}
	require.Len(t, channel.requested, carPeerBackoffHardFailures)
	p.carFetchOutcome.drain()

	p.fetchCARs(context.Background(), "creator", "", []cid.Cid{head.Cid()})

	require.Len(t, channel.requested, carPeerBackoffHardFailures, "a known-bad peer is not asked")
	require.Equal(t, map[string]int64{"creatorBackoff": 1}, reasonMap(p.carFetchOutcome.drain()))
}

// A peer that answers is no longer held back by a failure it has recovered from.
func TestFetchCARs_ForgetsBackoffOnceThePeerAnswers(t *testing.T) {
	head := compositeBlock(t, 1)
	channel := &fakeCARChannel{}
	p := fetchFixture(channel)

	p.fetchCARs(context.Background(), "creator", "", []cid.Cid{head.Cid()})
	require.Equal(t, carPeerDeprioritised, p.carPeerAvailability("creator"))

	channel.replies = map[string][]protocol.CARReply{
		"creator": {{CARs: [][]byte{carWith(t, head.Cid(), head)}}},
	}
	p.fetchCARs(context.Background(), "creator", "", []cid.Cid{head.Cid()})

	require.Equal(t, carPeerReady, p.carPeerAvailability("creator"))
}

// The window grows with consecutive failures and stops at the cap.
func TestBackOffCARPeer_GrowsTheWindowAndCaps(t *testing.T) {
	p := withReasonMaps(&P2P{})

	p.backOffCARPeer("peer")
	first := time.Until(p.carBackoff.peers["peer"].until)
	require.Greater(t, first, time.Duration(0))
	require.LessOrEqual(t, first, carPeerBackoffBase)

	p.backOffCARPeer("peer")
	second := time.Until(p.carBackoff.peers["peer"].until)
	require.Greater(t, second, carPeerBackoffBase)

	for range 20 {
		p.backOffCARPeer("peer")
	}
	require.LessOrEqual(t, time.Until(p.carBackoff.peers["peer"].until), carPeerBackoffMax)
	require.Equal(t, carPeerPassedOver, p.carPeerAvailability("peer"))
}

// An expired window is forgotten rather than left to accumulate failures.
func TestCARPeerAvailability_ForgetsAnExpiredWindow(t *testing.T) {
	p := withReasonMaps(&P2P{})
	p.backOffCARPeer("peer")
	p.carBackoff.peers["peer"].until = time.Now().Add(-time.Second)

	require.Equal(t, carPeerReady, p.carPeerAvailability("peer"))
	require.NotContains(t, p.carBackoff.peers, "peer")
}

// This node giving up on a message is not the peer failing, and must not count against it.
func TestFetchCARs_DoesNotBackOffAPeerWhenThisNodeCancels(t *testing.T) {
	head := compositeBlock(t, 1)
	p := fetchFixture(&fakeCARChannel{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p.fetchCARs(ctx, "creator", "", []cid.Cid{head.Cid()})

	require.Equal(t, map[string]int64{"creatorCanceled": 1}, reasonMap(p.carFetchOutcome.drain()))
	require.Equal(t, carPeerReady, p.carPeerAvailability("creator"))
}

func TestCARRequestFailure(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	require.Equal(t, fetchCanceled, carRequestFailure(cancelled, errors.New("any")))
	require.Equal(t, fetchTimeout,
		carRequestFailure(context.Background(), context.DeadlineExceeded))
	require.Equal(t, fetchResourceLimited,
		carRequestFailure(context.Background(), errors.New("stream reset by remote, error code: 4098")))
	require.Equal(t, fetchDialFailed,
		carRequestFailure(context.Background(), errors.New("failed to dial peer")))
	require.Equal(t, fetchRequestFailed,
		carRequestFailure(context.Background(), errors.New("stream reset")))
}

func TestCARFetchFailedPeer(t *testing.T) {
	// A resource limit counts against the peer: it is asking to be left alone for a moment.
	for _, outcome := range []string{
		fetchRequestFailed, fetchTimeout, fetchDialFailed, fetchResourceLimited,
	} {
		require.True(t, carFetchFailedPeer(outcome), outcome)
	}
	// A slot this node could not get is its own throttle, not the peer's.
	for _, outcome := range []string{
		fetchCanceled, fetchQueueFull, fetchBadReply, fetchNotServed, fetchFetched,
	} {
		require.False(t, carFetchFailedPeer(outcome), outcome)
	}
}

// The reset a peer's resource manager sends is recognised in both the renderings it reaches this
// node in, so it is never mistaken for an ordinary request failure.
func TestCARStreamResourceLimited(t *testing.T) {
	require.True(t, carStreamResourceLimited(
		errors.New("stream reset by remote, error code: 4098")))
	require.True(t, carStreamResourceLimited(
		errors.New("stream reset (remote): code: 0x1002")))

	require.False(t, carStreamResourceLimited(errors.New("stream reset")))
	require.False(t, carStreamResourceLimited(
		errors.New("stream reset by remote, error code: 4097")))
}

// A peer's slots are handed out up to the limit and no further, so this node never opens more
// streams to one peer than its resource manager will accept.
func TestAcquireCARSlot_HandsOutUpToTheLimit(t *testing.T) {
	p := withReasonMaps(&P2P{})

	releases := make([]func(), 0, maxInFlightCARRequestsPerPeer)
	for range maxInFlightCARRequestsPerPeer {
		release, outcome := p.acquireCARSlot(context.Background(), "peer")
		require.Empty(t, outcome)
		releases = append(releases, release)
	}

	// Another peer is limited separately.
	release, outcome := p.acquireCARSlot(context.Background(), "other")
	require.Empty(t, outcome)
	release()

	for _, release := range releases {
		release()
	}
}

// A request past the limit waits rather than being refused, and proceeds as soon as one of the
// requests ahead of it finishes.
func TestAcquireCARSlot_WaitsForASlotToComeFree(t *testing.T) {
	p := withReasonMaps(&P2P{})

	var release func()
	for range maxInFlightCARRequestsPerPeer {
		var outcome string
		release, outcome = p.acquireCARSlot(context.Background(), "peer")
		require.Empty(t, outcome)
	}

	queued := make(chan string, 1)
	go func() {
		got, outcome := p.acquireCARSlot(context.Background(), "peer")
		if outcome == "" {
			got()
		}
		queued <- outcome
	}()

	select {
	case <-queued:
		t.Fatal("a request past the limit must wait, not proceed")
	case <-time.After(50 * time.Millisecond):
	}

	release()
	require.Empty(t, <-queued)
}

// This node giving up on the message releases a queued request, which is its own doing rather
// than the peer's and is not counted against it.
func TestAcquireCARSlot_ReturnsCanceledWhenThisNodeGivesUp(t *testing.T) {
	p := withReasonMaps(&P2P{})
	for range maxInFlightCARRequestsPerPeer {
		_, outcome := p.acquireCARSlot(context.Background(), "peer")
		require.Empty(t, outcome)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, outcome := p.acquireCARSlot(ctx, "peer")

	require.Equal(t, fetchCanceled, outcome)
	require.False(t, carFetchFailedPeer(outcome))
}

// When no peer answers, every head falls back to the walk.
func TestFetchCARs_FailedRequestLeavesEveryEntryNil(t *testing.T) {
	head := compositeBlock(t, 1)
	p := fetchFixture(&fakeCARChannel{})

	cars := p.fetchCARs(context.Background(), "creator", "creator", []cid.Cid{head.Cid()})

	require.Equal(t, [][]byte{nil}, cars)
	require.Equal(t, int64(1), p.statCARFetchMissed.Load())
}

// A reply that stops short is followed by a request for the heads past its end, so a batch is
// fetched whole however many heads one reply covers.
func TestFetchCARs_AsksAgainForHeadsPastAShortReply(t *testing.T) {
	first := compositeBlock(t, 1)
	second := compositeBlock(t, 2)
	third := compositeBlock(t, 3)

	channel := &fakeCARChannel{replies: map[string][]protocol.CARReply{"creator": {
		{CARs: [][]byte{carWith(t, first.Cid(), first)}},
		{CARs: [][]byte{carWith(t, second.Cid(), second), carWith(t, third.Cid(), third)}},
	}}}
	p := fetchFixture(channel)

	heads := []cid.Cid{first.Cid(), second.Cid(), third.Cid()}
	cars := p.fetchCARs(context.Background(), "creator", "creator", heads)

	require.Equal(t, []carCall{{peer: "creator", heads: 3}, {peer: "creator", heads: 2}}, channel.requested)
	for i, head := range heads {
		require.True(t, carRootIs(cars[i], head))
	}
	require.Equal(t, int64(3), p.statCARFetched.Load())
	require.Zero(t, p.statCARFetchMissed.Load())
}

// A reply answering nothing ends the asking of that peer, and the heads it left fall back to
// the walk.
func TestFetchCARs_StopsOnAReplyThatAnswersNothing(t *testing.T) {
	first := compositeBlock(t, 1)
	second := compositeBlock(t, 2)

	channel := &fakeCARChannel{replies: map[string][]protocol.CARReply{"creator": {
		{CARs: [][]byte{carWith(t, first.Cid(), first)}},
		{},
	}}}
	p := fetchFixture(channel)

	cars := p.fetchCARs(context.Background(), "creator", "creator", []cid.Cid{first.Cid(), second.Cid()})

	require.Equal(t, []carCall{{peer: "creator", heads: 2}, {peer: "creator", heads: 1}}, channel.requested)
	require.NotNil(t, cars[0])
	require.Nil(t, cars[1])
	require.Equal(t, int64(1), p.statCARFetched.Load())
	require.Equal(t, int64(1), p.statCARFetchMissed.Load())
	require.Equal(t, map[string]int64{
		"creatorFetched":  1,
		"creatorBadReply": 1,
	}, reasonMap(p.carFetchOutcome.drain()))
}

// Nothing is asked when there is nothing to fetch, or no peer other than this node to ask.
func TestFetchCARs_SkipsTheRequestWithoutHeadsOrPeer(t *testing.T) {
	head := compositeBlock(t, 1)
	channel := &fakeCARChannel{}
	p := fetchFixture(channel)

	require.Empty(t, p.fetchCARs(context.Background(), "creator", "relay", nil))
	require.Equal(t, [][]byte{nil}, p.fetchCARs(context.Background(), "", "", []cid.Cid{head.Cid()}))
	require.Equal(t, [][]byte{nil}, p.fetchCARs(context.Background(), "peerID", "peerID", []cid.Cid{head.Cid()}))
	require.Empty(t, channel.requested)
}

func TestCARSources(t *testing.T) {
	require.Equal(t,
		[]carSource{{peerID: "c", label: carSourceCreator}, {peerID: "r", label: carSourceRelay}},
		carSources("c", "r", "self"))
	// A relay that is the creator is asked once, as the creator.
	require.Equal(t, []carSource{{peerID: "c", label: carSourceCreator}}, carSources("c", "c", "self"))
	// An update naming no creator, as a peer predating Creator would send, asks the relay.
	require.Equal(t, []carSource{{peerID: "r", label: carSourceRelay}}, carSources("", "r", "self"))
	// This node is never asked.
	require.Equal(t, []carSource{{peerID: "r", label: carSourceRelay}}, carSources("self", "r", "self"))
}

func TestCARRootIs(t *testing.T) {
	head := compositeBlock(t, 1)
	other := compositeBlock(t, 2)
	data := carWith(t, head.Cid(), head)

	require.True(t, carRootIs(data, head.Cid()))
	require.False(t, carRootIs(data, other.Cid()))
	require.False(t, carRootIs([]byte("not a car"), head.Cid()))
}
