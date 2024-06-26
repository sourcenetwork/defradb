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

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
	mh "github.com/multiformats/go-multihash"
	badger "github.com/sourcenetwork/badger/v4"
	rpc "github.com/sourcenetwork/go-libp2p-pubsub-rpc"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/acp"
	"github.com/sourcenetwork/defradb/client"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/datastore/memory"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/db"
	netutils "github.com/sourcenetwork/defradb/net/utils"
)

func emptyBlock() []byte {
	block := coreblock.Block{
		Delta: crdt.CRDT{
			CompositeDAGDelta: &crdt.CompositeDAGDelta{},
		},
	}
	b, _ := block.Marshal()
	return b
}

func createCID(doc *client.Document) (cid.Cid, error) {
	pref := cid.V1Builder{
		Codec:    cid.DagCBOR,
		MhType:   mh.SHA2_256,
		MhLength: 0, // default length
	}

	buf, err := doc.Bytes()
	if err != nil {
		return cid.Cid{}, err
	}

	// And then feed it some data
	c, err := pref.Sum(buf)
	if err != nil {
		return cid.Cid{}, err
	}
	return c, nil
}

const randomMultiaddr = "/ip4/127.0.0.1/tcp/0"

func newTestPeer(ctx context.Context, t *testing.T) (client.DB, *Peer) {
	store := memory.NewDatastore(ctx)
	acpLocal := acp.NewLocalACP()
	acpLocal.Init(context.Background(), "")
	db, err := db.NewDB(ctx, store, immutable.Some[acp.ACP](acpLocal), nil)
	require.NoError(t, err)

	n, err := NewPeer(
		ctx,
		db.Rootstore(),
		db.Blockstore(),
		db.Events(),
		WithListenAddresses(randomMultiaddr),
	)
	require.NoError(t, err)

	return db, n
}

func TestNewPeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()
	p, err := NewPeer(ctx, db.Rootstore(), db.Blockstore(), db.Events())
	require.NoError(t, err)
	p.Close()
}

func TestNewPeer_NoDB_NilDBError(t *testing.T) {
	ctx := context.Background()
	_, err := NewPeer(ctx, nil, nil, nil)
	require.ErrorIs(t, err, ErrNilDB)
}

func TestStartAndClose_NoError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()

	err := p.Start()
	require.NoError(t, err)
}

func TestStart_WithKnownPeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db1, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db1.Close()

	store2 := memory.NewDatastore(ctx)
	db2, err := db.NewDB(ctx, store2, acp.NoACP, nil)
	require.NoError(t, err)
	defer db2.Close()

	n1, err := NewPeer(
		ctx,
		db1.Rootstore(),
		db1.Blockstore(),
		db1.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n1.Close()
	n2, err := NewPeer(
		ctx,
		db2.Rootstore(),
		db2.Blockstore(),
		db2.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n2.Close()

	addrs, err := netutils.ParsePeers([]string{n1.host.Addrs()[0].String() + "/p2p/" + n1.PeerID().String()})
	if err != nil {
		t.Fatal(err)
	}
	n2.Bootstrap(addrs)

	err = n2.Start()
	require.NoError(t, err)
}

func TestStart_WithOfflineKnownPeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db1, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db1.Close()

	store2 := memory.NewDatastore(ctx)
	db2, err := db.NewDB(ctx, store2, acp.NoACP, nil)
	require.NoError(t, err)
	defer db2.Close()

	n1, err := NewPeer(
		ctx,
		db1.Rootstore(),
		db1.Blockstore(),
		db1.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n1.Close()
	n2, err := NewPeer(
		ctx,
		db2.Rootstore(),
		db2.Blockstore(),
		db2.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n2.Close()

	addrs, err := netutils.ParsePeers([]string{n1.host.Addrs()[0].String() + "/p2p/" + n1.PeerID().String()})
	if err != nil {
		t.Fatal(err)
	}
	n2.Bootstrap(addrs)
	n1.Close()

	// give time for n1 to close
	time.Sleep(100 * time.Millisecond)

	err = n2.Start()
	require.NoError(t, err)
}

func TestRegisterNewDocument_NoError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Definition())
	require.NoError(t, err)

	cid, err := createCID(doc)
	require.NoError(t, err)

	err = p.RegisterNewDocument(ctx, doc.ID(), cid, emptyBlock(), col.SchemaRoot())
	require.NoError(t, err)
}

func TestRegisterNewDocument_RPCTopicAlreadyRegisteredError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Definition())
	require.NoError(t, err)

	_, err = rpc.NewTopic(ctx, p.ps, p.host.ID(), doc.ID().String(), true)
	require.NoError(t, err)

	cid, err := createCID(doc)
	require.NoError(t, err)

	err = p.RegisterNewDocument(ctx, doc.ID(), cid, emptyBlock(), col.SchemaRoot())
	require.Equal(t, err.Error(), "creating topic: joining topic: topic already exists")
}

func TestHandleDocCreateLog_NoError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()

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

	headCID, err := getHead(ctx, db, doc.ID())
	require.NoError(t, err)

	b, err := db.Blockstore().AsIPLDStorage().Get(ctx, headCID.KeyString())
	require.NoError(t, err)

	err = p.handleDocCreateLog(event.Update{
		DocID:      doc.ID().String(),
		Cid:        headCID,
		SchemaRoot: col.SchemaRoot(),
		Block:      b,
	})
	require.NoError(t, err)
}

func TestHandleDocCreateLog_WithInvalidDocID_NoError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()

	err := p.handleDocCreateLog(event.Update{
		DocID: "some-invalid-key",
	})
	require.ErrorContains(t, err, "failed to get DocID from broadcast message: selected encoding not supported")
}

func TestHandleDocCreateLog_WithExistingTopic_TopicExistsError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()

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

	_, err = rpc.NewTopic(ctx, p.ps, p.host.ID(), doc.ID().String(), true)
	require.NoError(t, err)

	err = p.handleDocCreateLog(event.Update{
		DocID:      doc.ID().String(),
		SchemaRoot: col.SchemaRoot(),
	})
	require.ErrorContains(t, err, "topic already exists")
}

func TestHandleDocUpdateLog_NoError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()

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

	headCID, err := getHead(ctx, db, doc.ID())
	require.NoError(t, err)

	b, err := db.Blockstore().AsIPLDStorage().Get(ctx, headCID.KeyString())
	require.NoError(t, err)

	err = p.handleDocUpdateLog(event.Update{
		DocID:      doc.ID().String(),
		Cid:        headCID,
		SchemaRoot: col.SchemaRoot(),
		Block:      b,
	})
	require.NoError(t, err)
}

func TestHandleDoUpdateLog_WithInvalidDocID_NoError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()

	err := p.handleDocUpdateLog(event.Update{
		DocID: "some-invalid-key",
	})
	require.ErrorContains(t, err, "failed to get DocID from broadcast message: selected encoding not supported")
}

func TestHandleDocUpdateLog_WithExistingDocIDTopic_TopicExistsError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()

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

	headCID, err := getHead(ctx, db, doc.ID())
	require.NoError(t, err)

	b, err := db.Blockstore().AsIPLDStorage().Get(ctx, headCID.KeyString())
	require.NoError(t, err)

	_, err = rpc.NewTopic(ctx, p.ps, p.host.ID(), doc.ID().String(), true)
	require.NoError(t, err)

	err = p.handleDocUpdateLog(event.Update{
		DocID:      doc.ID().String(),
		Cid:        headCID,
		SchemaRoot: col.SchemaRoot(),
		Block:      b,
	})
	require.ErrorContains(t, err, "topic already exists")
}

func TestHandleDocUpdateLog_WithExistingSchemaTopic_TopicExistsError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()

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

	headCID, err := getHead(ctx, db, doc.ID())
	require.NoError(t, err)

	b, err := db.Blockstore().AsIPLDStorage().Get(ctx, headCID.KeyString())
	require.NoError(t, err)

	_, err = rpc.NewTopic(ctx, p.ps, p.host.ID(), col.SchemaRoot(), true)
	require.NoError(t, err)

	err = p.handleDocUpdateLog(event.Update{
		DocID:      doc.ID().String(),
		Cid:        headCID,
		SchemaRoot: col.SchemaRoot(),
		Block:      b,
	})
	require.ErrorContains(t, err, "topic already exists")
}

func FixtureNewMemoryDBWithBroadcaster(t *testing.T) client.DB {
	var database client.DB
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	require.NoError(t, err)
	database, err = db.NewDB(ctx, rootstore, acp.NoACP, nil)
	require.NoError(t, err)
	return database
}

func TestNewPeer_WithEnableRelay_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()
	n, err := NewPeer(
		context.Background(),
		db.Rootstore(),
		db.Blockstore(),
		db.Events(),
		WithEnableRelay(true),
	)
	require.NoError(t, err)
	n.Close()
}

func TestNewPeer_WithDBClosed_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)

	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	db.Close()

	_, err = NewPeer(
		context.Background(),
		db.Rootstore(),
		db.Blockstore(),
		db.Events(),
	)
	require.ErrorContains(t, err, "datastore closed")
}

func TestNewPeer_NoPubSub_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n, err := NewPeer(
		context.Background(),
		db.Rootstore(),
		db.Blockstore(),
		db.Events(),
		WithEnablePubSub(false),
	)
	require.NoError(t, err)
	require.Nil(t, n.ps)
	n.Close()
}

func TestNewPeer_WithEnablePubSub_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n, err := NewPeer(
		ctx,
		db.Rootstore(),
		db.Blockstore(),
		db.Events(),
		WithEnablePubSub(true),
	)

	require.NoError(t, err)
	// overly simple check of validity of pubsub, avoiding the process of creating a PubSub
	require.NotNil(t, n.ps)
	n.Close()
}

func TestNodeClose_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()
	n, err := NewPeer(
		context.Background(),
		db.Rootstore(),
		db.Blockstore(),
		db.Events(),
	)
	require.NoError(t, err)
	n.Close()
}

func TestNewPeer_BootstrapWithNoPeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n1, err := NewPeer(
		ctx,
		db.Rootstore(),
		db.Blockstore(),
		db.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	n1.Bootstrap([]peer.AddrInfo{})
	n1.Close()
}

func TestNewPeer_BootstrapWithOnePeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()
	n1, err := NewPeer(
		ctx,
		db.Rootstore(),
		db.Blockstore(),
		db.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n1.Close()
	n2, err := NewPeer(
		ctx,
		db.Rootstore(),
		db.Blockstore(),
		db.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n2.Close()
	addrs, err := netutils.ParsePeers([]string{n1.host.Addrs()[0].String() + "/p2p/" + n1.PeerID().String()})
	if err != nil {
		t.Fatal(err)
	}
	n2.Bootstrap(addrs)
}

func TestNewPeer_BootstrapWithOneValidPeerAndManyInvalidPeers_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n1, err := NewPeer(
		ctx,
		db.Rootstore(),
		db.Blockstore(),
		db.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n1.Close()
	n2, err := NewPeer(
		ctx,
		db.Rootstore(),
		db.Blockstore(),
		db.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	defer n2.Close()
	addrs, err := netutils.ParsePeers([]string{
		n1.host.Addrs()[0].String() + "/p2p/" + n1.PeerID().String(),
		"/ip4/0.0.0.0/tcp/1234/p2p/" + "12D3KooWC8YY6Tx3uAeHsdBmoy7PJPwqXAHE4HkCZ5veankKWci6",
		"/ip4/0.0.0.0/tcp/1235/p2p/" + "12D3KooWC8YY6Tx3uAeHsdBmoy7PJPwqXAHE4HkCZ5veankKWci5",
		"/ip4/0.0.0.0/tcp/1236/p2p/" + "12D3KooWC8YY6Tx3uAeHsdBmoy7PJPwqXAHE4HkCZ5veankKWci4",
	})
	require.NoError(t, err)
	n2.Bootstrap(addrs)
}

func TestListenAddrs_WithListenAddresses_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n, err := NewPeer(
		context.Background(),
		db.Rootstore(),
		db.Blockstore(),
		db.Events(),
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	require.Contains(t, n.ListenAddrs()[0].String(), "/tcp/")
	n.Close()
}
