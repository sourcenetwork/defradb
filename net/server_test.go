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
	"time"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/stretchr/testify/require"
	rpc "github.com/textileio/go-libp2p-pubsub-rpc"
	grpcpeer "google.golang.org/grpc/peer"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore/memory"
	"github.com/sourcenetwork/defradb/errors"
	net_pb "github.com/sourcenetwork/defradb/net/pb"
)

func TestNewServerSimple(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	_, err := newServer(n.Peer, db)
	require.NoError(t, err)
}

func TestNewServerWithDBClosed(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	db.Close()

	_, err := newServer(n.Peer, db)
	require.ErrorIs(t, err, memory.ErrClosed)
}

var mockError = errors.New("mock error")

type mockDBColError struct {
	client.DB
}

func (mDB *mockDBColError) GetAllCollections(context.Context) ([]client.Collection, error) {
	return nil, mockError
}

func TestNewServerWithGetAllCollectionError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	mDB := mockDBColError{db}
	_, err := newServer(n.Peer, &mDB)
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

	_, err = newServer(n.Peer, db)
	require.NoError(t, err)
}

type mockDBDockeysError struct {
	client.DB
}

func (mDB *mockDBDockeysError) GetAllCollections(context.Context) ([]client.Collection, error) {
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
func (mCol *mockCollection) GetAllDocKeys(ctx context.Context) (<-chan client.DocKeysResult, error) {
	return nil, mockError
}

func TestNewServerWithGetAllDockeysError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	mDB := mockDBDockeysError{db}

	_, err = newServer(n.Peer, &mDB)
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

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`))
	require.NoError(t, err)

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	_, err = rpc.NewTopic(ctx, n.Peer.ps, n.Peer.host.ID(), doc.Key().String(), true)
	require.NoError(t, err)

	_, err = newServer(n.Peer, db)
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

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`))
	require.NoError(t, err)

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	n.Peer.host = &mockHost{n.Peer.host}

	_, err = newServer(n.Peer, db)
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

func TestDocQueue(t *testing.T) {
	q := docQueue{
		docs: make(map[string]chan struct{}),
	}

	testKey := "test"

	q.add(testKey)
	go q.add(testKey)
	// give time for the goroutine to block
	time.Sleep(10 * time.Millisecond)
	require.Len(t, q.docs, 1)
	q.done(testKey)
	// give time for the goroutine to add the key
	time.Sleep(10 * time.Millisecond)
	q.mu.Lock()
	require.Len(t, q.docs, 1)
	q.mu.Unlock()
	q.done(testKey)
	q.mu.Lock()
	require.Len(t, q.docs, 0)
	q.mu.Unlock()
}

func TestPushLog(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`))
	require.NoError(t, err)

	cid, err := createCID(doc)
	require.NoError(t, err)

	ctx = grpcpeer.NewContext(ctx, &grpcpeer.Peer{
		Addr: addr{n.PeerID()},
	})

	block := &EmptyNode{}

	_, err = n.server.PushLog(ctx, &net_pb.PushLogRequest{
		Body: &net_pb.PushLogRequest_Body{
			DocKey:     []byte(doc.Key().String()),
			Cid:        cid.Bytes(),
			SchemaRoot: []byte(col.SchemaRoot()),
			Creator:    n.PeerID().String(),
			Log: &net_pb.Document_Log{
				Block: block.RawData(),
			},
		},
	})
	require.NoError(t, err)
}
