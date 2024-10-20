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
	db, err := db.NewDB(
		ctx,
		store,
		immutable.Some[acp.ACP](acpLocal),
		nil,
		db.WithRetryInterval([]time.Duration{time.Second}),
	)
	require.NoError(t, err)

	n, err := NewPeer(
		ctx,
		db.Blockstore(),
		db.Encstore(),
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
	p, err := NewPeer(ctx, db.Blockstore(), db.Encstore(), db.Events())
	require.NoError(t, err)
	p.Close()
}

func TestNewPeer_NoDB_NilDBError(t *testing.T) {
	ctx := context.Background()
	_, err := NewPeer(ctx, nil, nil, nil, nil)
	require.ErrorIs(t, err, ErrNilDB)
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
		db1.Blockstore(),
		db1.Encstore(),
		db1.Events(),
		WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)
	defer n1.Close()
	n2, err := NewPeer(
		ctx,
		db2.Blockstore(),
		db1.Encstore(),
		db2.Events(),
		WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)
	defer n2.Close()

	err = n2.Connect(ctx, n1.PeerInfo())
	require.NoError(t, err)
}

func TestHandleLog_NoError(t *testing.T) {
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

	err = p.handleLog(event.Update{
		DocID:      doc.ID().String(),
		Cid:        headCID,
		SchemaRoot: col.SchemaRoot(),
		Block:      b,
	})
	require.NoError(t, err)
}

func TestHandleLog_WithInvalidDocID_NoError(t *testing.T) {
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()

	err := p.handleLog(event.Update{
		DocID: "some-invalid-key",
	})
	require.ErrorContains(t, err, "failed to get DocID from broadcast message: selected encoding not supported")
}

func TestHandleLog_WithExistingTopic_TopicExistsError(t *testing.T) {
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

	_, err = rpc.NewTopic(ctx, p.ps, p.host.ID(), "bae-7fca96a2-5f01-5558-a81f-09b47587f26d", true)
	require.NoError(t, err)

	err = p.handleLog(event.Update{
		DocID:      doc.ID().String(),
		SchemaRoot: col.SchemaRoot(),
	})
	require.ErrorContains(t, err, "topic already exists")
}

func TestHandleLog_WithExistingSchemaTopic_TopicExistsError(t *testing.T) {
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

	_, err = rpc.NewTopic(ctx, p.ps, p.host.ID(), "bafkreia7ljiy5oief4dp5xsk7t7zlgfjzqh3537hw7rtttjzchybfxtn4u", true)
	require.NoError(t, err)

	err = p.handleLog(event.Update{
		DocID:      doc.ID().String(),
		Cid:        cid,
		SchemaRoot: col.SchemaRoot(),
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
		db.Blockstore(),
		db.Encstore(),
		db.Events(),
		WithEnableRelay(true),
	)
	require.NoError(t, err)
	n.Close()
}

func TestNewPeer_NoPubSub_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n, err := NewPeer(
		context.Background(),
		db.Blockstore(),
		db.Encstore(),
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
		db.Blockstore(),
		db.Encstore(),
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
		db.Blockstore(),
		db.Encstore(),
		db.Events(),
	)
	require.NoError(t, err)
	n.Close()
}

func TestListenAddrs_WithListenAddresses_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n, err := NewPeer(
		context.Background(),
		db.Blockstore(),
		db.Encstore(),
		db.Events(),
		WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)
	require.Contains(t, n.ListenAddrs()[0].String(), "/tcp/")
	n.Close()
}

func TestPeer_WithBootstrapPeers_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, acp.NoACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n, err := NewPeer(
		context.Background(),
		db.Blockstore(),
		db.Encstore(),
		db.Events(),
		WithBootstrapPeers("/ip4/127.0.0.1/tcp/6666/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ"),
	)
	require.NoError(t, err)

	n.Close()
}
