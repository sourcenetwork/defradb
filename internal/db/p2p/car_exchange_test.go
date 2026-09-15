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

// fakeCARChannel answers CAR requests with its replies in turn, and fails once they run out or
// when err is set. requested records how many heads each request asked for.
type fakeCARChannel struct {
	replies   []protocol.CARReply
	err       error
	requested []int
}

func (f *fakeCARChannel) SendRequest(
	_ context.Context,
	req protocol.CARRequest,
	_ string,
) (protocol.CARReply, error) {
	f.requested = append(f.requested, len(req.CIDs))
	if f.err != nil {
		return protocol.CARReply{}, f.err
	}
	if len(f.requested) > len(f.replies) {
		return protocol.CARReply{}, errors.New("no reply left")
	}
	return f.replies[len(f.requested)-1], nil
}

// A held head is served as a CAR rooted at it. A head this node does not hold, or bytes that
// are not a CID, get an empty entry without costing the requester the rest of the reply.
func TestServeCARs_ServesHeldHeadsAndLeavesTheRestEmpty(t *testing.T) {
	child := undecodableBlock(t, "field block")
	_, root := compositeLinking(t, child.Cid())
	p := carFixture(t, root, child)
	p.db = servingDB{rootstoreDB: p.db.(rootstoreDB)}

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
}

// A request for more heads than a batch holds is answered only up to the cap.
func TestServeCARs_AnswersAtMostTheCap(t *testing.T) {
	p := carFixture(t)
	p.db = servingDB{rootstoreDB: p.db.(rootstoreDB)}

	absent := compositeBlock(t, 1)
	req := make([][]byte, maxCARRequestCIDs+5)
	for i := range req {
		req[i] = absent.Cid().Bytes()
	}

	reply := p.serveCARs(context.Background(), "peer", req)
	require.Len(t, reply.CARs, maxCARRequestCIDs)
}

// A served CAR is kept only when rooted at the head it was asked for. One rooted elsewhere, or
// not served, leaves the entry nil so the caller walks the DAG for it.
func TestFetchCARs_KeepsOnlyCARsRootedAtTheirHead(t *testing.T) {
	_, first := compositeLinking(t)
	second := compositeBlock(t, 2)
	third := compositeBlock(t, 3)

	channel := &fakeCARChannel{replies: []protocol.CARReply{{CARs: [][]byte{
		carWith(t, first.Cid(), first),
		carWith(t, third.Cid(), third), // rooted at the wrong head
		nil,
	}}}}
	p := withReasonMaps(&P2P{host: &SimpleMockHost{}, carProtocol: channel})

	cars := p.fetchCARs(context.Background(), "peer", []cid.Cid{first.Cid(), second.Cid(), third.Cid()})

	require.Len(t, cars, 3)
	require.NotNil(t, cars[0])
	require.Nil(t, cars[1])
	require.Nil(t, cars[2])
	require.Equal(t, int64(1), p.statCARFetched.Load())
	require.Equal(t, int64(2), p.statCARFetchMissed.Load())
}

// A peer that predates the exchange fails the request, and every head falls back to the walk.
func TestFetchCARs_FailedRequestLeavesEveryEntryNil(t *testing.T) {
	head := compositeBlock(t, 1)
	channel := &fakeCARChannel{err: errors.New("protocol not supported")}
	p := withReasonMaps(&P2P{host: &SimpleMockHost{}, carProtocol: channel})

	cars := p.fetchCARs(context.Background(), "peer", []cid.Cid{head.Cid()})

	require.Equal(t, [][]byte{nil}, cars)
	require.Equal(t, int64(1), p.statCARFetchMissed.Load())
}

// A reply that stops short is followed by a request for the heads past its end, so a batch is
// fetched whole however many heads one reply covers.
func TestFetchCARs_AsksAgainForHeadsPastAShortReply(t *testing.T) {
	first := compositeBlock(t, 1)
	second := compositeBlock(t, 2)
	third := compositeBlock(t, 3)

	channel := &fakeCARChannel{replies: []protocol.CARReply{
		{CARs: [][]byte{carWith(t, first.Cid(), first)}},
		{CARs: [][]byte{carWith(t, second.Cid(), second), carWith(t, third.Cid(), third)}},
	}}
	p := withReasonMaps(&P2P{host: &SimpleMockHost{}, carProtocol: channel})

	cars := p.fetchCARs(context.Background(), "peer", []cid.Cid{first.Cid(), second.Cid(), third.Cid()})

	require.Equal(t, []int{3, 2}, channel.requested)
	for i, head := range []cid.Cid{first.Cid(), second.Cid(), third.Cid()} {
		require.True(t, carRootIs(cars[i], head))
	}
	require.Equal(t, int64(3), p.statCARFetched.Load())
	require.Zero(t, p.statCARFetchMissed.Load())
}

// A reply answering nothing ends the fetch, and the heads it left fall back to the walk.
func TestFetchCARs_StopsOnAReplyThatAnswersNothing(t *testing.T) {
	first := compositeBlock(t, 1)
	second := compositeBlock(t, 2)

	channel := &fakeCARChannel{replies: []protocol.CARReply{
		{CARs: [][]byte{carWith(t, first.Cid(), first)}},
		{},
	}}
	p := withReasonMaps(&P2P{host: &SimpleMockHost{}, carProtocol: channel})

	cars := p.fetchCARs(context.Background(), "peer", []cid.Cid{first.Cid(), second.Cid()})

	require.Equal(t, []int{2, 1}, channel.requested)
	require.NotNil(t, cars[0])
	require.Nil(t, cars[1])
	require.Equal(t, int64(1), p.statCARFetched.Load())
	require.Equal(t, int64(1), p.statCARFetchMissed.Load())
}

// A batch larger than the old per-message size is served in one reply, not cut to 64.
func TestServeCARs_AnswersABatchLargerThanTheSenderBatchSize(t *testing.T) {
	child := undecodableBlock(t, "field block")
	_, root := compositeLinking(t, child.Cid())
	p := carFixture(t, root, child)
	p.db = servingDB{rootstoreDB: p.db.(rootstoreDB)}
	p.carCache = newCARCache(carCacheMaxBytes)

	req := make([][]byte, 4*batchMaxDocs)
	for i := range req {
		req[i] = root.Cid().Bytes()
	}

	reply := p.serveCARs(context.Background(), "peer", req)
	require.Len(t, reply.CARs, len(req))
	require.True(t, carRootIs(reply.CARs[len(req)-1], root.Cid()))
}

// Nothing is asked of a peer when there is nothing to fetch or no peer to ask.
func TestFetchCARs_SkipsTheRequestWithoutHeadsOrPeer(t *testing.T) {
	head := compositeBlock(t, 1)
	channel := &fakeCARChannel{}
	p := withReasonMaps(&P2P{host: &SimpleMockHost{}, carProtocol: channel})

	require.Empty(t, p.fetchCARs(context.Background(), "peer", nil))
	require.Equal(t, [][]byte{nil}, p.fetchCARs(context.Background(), "", []cid.Cid{head.Cid()}))
	require.Equal(t, [][]byte{nil}, p.fetchCARs(context.Background(), "peerID", []cid.Cid{head.Cid()}))
	require.Empty(t, channel.requested)
}

func TestCARRootIs(t *testing.T) {
	head := compositeBlock(t, 1)
	other := compositeBlock(t, 2)
	data := carWith(t, head.Cid(), head)

	require.True(t, carRootIs(data, head.Cid()))
	require.False(t, carRootIs(data, other.Cid()))
	require.False(t, carRootIs([]byte("not a car"), head.Cid()))
}
