// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"context"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore/query"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	rpc "github.com/sourcenetwork/go-libp2p-pubsub-rpc"
	"github.com/stretchr/testify/require"
	grpcpeer "google.golang.org/grpc/peer"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore/memory"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	net_pb "github.com/sourcenetwork/defradb/net/pb"
)

func TestNewServerSimple(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	_, err := newServer(n.Peer)
	require.NoError(t, err)
}

func TestNewServerWithDBClosed(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	db.Close()

	_, err := newServer(n.Peer)
	require.ErrorIs(t, err, memory.ErrClosed)
}

var mockError = errors.New("mock error")

type mockDBColError struct {
	client.DB
}

func (mDB *mockDBColError) GetCollections(context.Context, client.CollectionFetchOptions) ([]client.Collection, error) {
	return nil, mockError
}

func TestNewServerWithGetAllCollectionError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	mDB := mockDBColError{db}
	n.Peer.db = &mDB
	_, err := newServer(n.Peer)
	require.ErrorIs(t, err, mockError)
}

func TestNewServerWithCollectionSubscribed(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	err = n.AddP2PCollections(ctx, []string{col.SchemaRoot()})
	require.NoError(t, err)

	_, err = newServer(n.Peer)
	require.NoError(t, err)
}

type mockDBDocIDsError struct {
	client.DB
}

func (mDB *mockDBDocIDsError) GetCollections(context.Context, client.CollectionFetchOptions) ([]client.Collection, error) {
	return []client.Collection{
		&mockCollection{},
	}, nil
}

type mockCollection struct {
	client.Collection
}

func (mCol *mockCollection) SchemaRoot() string {
	return "mockColID"
}
func (mCol *mockCollection) GetAllDocIDs(
	ctx context.Context,
) (<-chan client.DocIDResult, error) {
	return nil, mockError
}

func TestNewServerWithGetAllDocIDsError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	mDB := mockDBDocIDsError{db}
	n.Peer.db = &mDB
	_, err = newServer(n.Peer)
	require.ErrorIs(t, err, mockError)
}

func TestNewServerWithAddTopicError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Definition())
	require.NoError(t, err)

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	_, err = rpc.NewTopic(ctx, n.Peer.ps, n.Peer.host.ID(), doc.ID().String(), true)
	require.NoError(t, err)

	_, err = newServer(n.Peer)
	require.ErrorContains(t, err, "topic already exists")
}

type mockHost struct {
	host.Host
}

func (mH *mockHost) EventBus() event.Bus {
	return &mockBus{}
}

type mockBus struct {
	event.Bus
}

func (mB *mockBus) Emitter(eventType any, opts ...event.EmitterOpt) (event.Emitter, error) {
	return nil, mockError
}

func (mB *mockBus) Subscribe(eventType any, opts ...event.SubscriptionOpt) (event.Subscription, error) {
	return nil, mockError
}

func TestNewServerWithEmitterError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Definition())
	require.NoError(t, err)

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	n.Peer.host = &mockHost{n.Peer.host}

	_, err = newServer(n.Peer)
	require.NoError(t, err)
}

func TestGetDocGraph(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	r, err := n.server.GetDocGraph(ctx, &net_pb.GetDocGraphRequest{})
	require.Nil(t, r)
	require.Nil(t, err)
}

func TestPushDocGraph(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	r, err := n.server.PushDocGraph(ctx, &net_pb.PushDocGraphRequest{})
	require.Nil(t, r)
	require.Nil(t, err)
}

func TestGetLog(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	r, err := n.server.GetLog(ctx, &net_pb.GetLogRequest{})
	require.Nil(t, r)
	require.Nil(t, err)
}

func TestGetHeadLog(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	r, err := n.server.GetHeadLog(ctx, &net_pb.GetHeadLogRequest{})
	require.Nil(t, r)
	require.Nil(t, err)
}

func getHead(ctx context.Context, db client.DB, docID client.DocID) (cid.Cid, error) {
	prefix := core.DataStoreKeyFromDocID(docID).ToHeadStoreKey().WithFieldId(core.COMPOSITE_NAMESPACE).ToString()
	results, err := db.Headstore().Query(ctx, query.Query{Prefix: prefix})
	if err != nil {
		return cid.Undef, err
	}
	entries, err := results.Rest()
	if err != nil {
		return cid.Undef, err
	}

	if len(entries) > 0 {
		hsKey, err := core.NewHeadStoreKey(entries[0].Key)
		if err != nil {
			return cid.Undef, err
		}
		return hsKey.Cid, nil
	}
	return cid.Undef, errors.New("no head found")
}

func TestPushLog(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	err := n.Start()
	require.NoError(t, err)

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Definition())
	require.NoError(t, err)

	ctx = grpcpeer.NewContext(ctx, &grpcpeer.Peer{
		Addr: addr{n.PeerID()},
	})

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	headCID, err := getHead(ctx, db, doc.ID())
	require.NoError(t, err)

	b, err := db.Blockstore().AsIPLDStorage().Get(ctx, headCID.KeyString())
	require.NoError(t, err)

	_, err = n.server.PushLog(ctx, &net_pb.PushLogRequest{
		Body: &net_pb.PushLogRequest_Body{
			DocID:      []byte(doc.ID().String()),
			Cid:        headCID.Bytes(),
			SchemaRoot: []byte(col.SchemaRoot()),
			Creator:    n.PeerID().String(),
			Log: &net_pb.Document_Log{
				Block: b,
			},
		},
	})
	require.NoError(t, err)
}
