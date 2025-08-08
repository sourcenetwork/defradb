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

	badgerds "github.com/dgraph-io/badger/v4"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corekv/badger"
	"github.com/sourcenetwork/corekv/memory"
	rpc "github.com/sourcenetwork/go-libp2p-pubsub-rpc"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/net/config"
)

func emptyBlock() []byte {
	block := coreblock.Block{
		Delta: crdt.CRDT{
			DocCompositeDelta: &crdt.DocCompositeDelta{},
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

type testdb interface {
	client.TxnStore
	Rootstore() corekv.TxnStore
	Close()
}

func newTestPeer(ctx context.Context, t *testing.T) (testdb, *Peer) {
	store, err := badger.NewDatastore("", badgerds.DefaultOptions("").WithInMemory(true))
	require.NoError(t, err)

	adminInfo, err := db.NewNACInfo(ctx, "", false)
	require.NoError(t, err)

	localDocumentACP, err := dac.NewLocalDocumentACP("")
	require.NoError(t, err)
	db, err := db.NewDB(
		ctx,
		store,
		adminInfo,
		immutable.Some(localDocumentACP),
		nil,
	)
	require.NoError(t, err)

	n, err := NewPeer(
		ctx,
		db.Events(),
		immutable.None[dac.DocumentACP](),
		db,
		config.WithListenAddresses(randomMultiaddr),
		config.WithRetryInterval([]time.Duration{time.Second}),
	)
	require.NoError(t, err)

	return db, n
}

func TestNewPeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	adminInfo, err := db.NewNACInfo(ctx, "", false)
	require.NoError(t, err)
	db, err := db.NewDB(ctx, store, adminInfo, dac.NoDocumentACP, nil)
	require.NoError(t, err)
	defer db.Close()
	p, err := NewPeer(ctx, db.Events(), immutable.None[dac.DocumentACP](), db)
	require.NoError(t, err)
	p.Close()
}

func TestNewPeer_NoDB_NilDBError(t *testing.T) {
	ctx := context.Background()
	_, err := NewPeer(ctx, nil, immutable.None[dac.DocumentACP](), nil, nil)
	require.ErrorIs(t, err, ErrNilDB)
}

func TestStart_WithKnownPeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	adminInfo, err := db.NewNACInfo(ctx, "", false)
	require.NoError(t, err)
	db1, err := db.NewDB(ctx, store, adminInfo, dac.NoDocumentACP, nil)
	require.NoError(t, err)
	defer db1.Close()

	require.NoError(t, err)
	store2 := memory.NewDatastore(ctx)
	db2, err := db.NewDB(ctx, store2, adminInfo, dac.NoDocumentACP, nil)
	require.NoError(t, err)
	defer db2.Close()

	n1, err := NewPeer(
		ctx,
		db1.Events(),
		immutable.None[dac.DocumentACP](),
		db1,
		config.WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)
	defer n1.Close()
	n2, err := NewPeer(
		ctx,
		db2.Events(),
		immutable.None[dac.DocumentACP](),
		db2,
		config.WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)
	defer n2.Close()

	err = n2.Connect(ctx, n1.PeerInfo())
	require.NoError(t, err)
}

func TestHandleLog_NoError(t *testing.T) {
	docID := "bae-7fca96a2-5f01-5558-a81f-09b47587f26d"
	collectionID := "bafkreia7ljiy5oief4dp5xsk7t7zlgfjzqh3537hw7rtttjzchybfxtn4u"
	ctx := context.Background()
	db, p := newTestPeer(ctx, t)
	defer db.Close()
	defer p.Close()

	err := p.handleLog(event.Update{
		DocID:        docID,
		CollectionID: collectionID,
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

	_, err = rpc.NewTopic(ctx, p.ps, p.host.ID(), "bae-a911f9cc-217a-58a3-a2f4-96548197403e", true)
	require.NoError(t, err)

	err = p.handleLog(event.Update{
		DocID:        doc.ID().String(),
		CollectionID: col.Version().CollectionID,
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

	_, err = rpc.NewTopic(ctx, p.ps, p.host.ID(), "bafyreib6hugraqnvqu25yamseuzztgnq24kepo7ysvjkpbh2eaag6jj3mm", true)
	require.NoError(t, err)

	err = p.handleLog(event.Update{
		DocID:        doc.ID().String(),
		Cid:          cid,
		CollectionID: col.Version().CollectionID,
	})
	require.ErrorContains(t, err, "topic already exists")
}

func newTestDB(ctx context.Context, t *testing.T) *db.DB {
	rootstore := memory.NewDatastore(ctx)
	adminInfo, err := db.NewNACInfo(ctx, "", false)
	require.NoError(t, err)
	database, err := db.NewDB(ctx, rootstore, adminInfo, dac.NoDocumentACP, nil)
	require.NoError(t, err)
	return database
}

func TestNewPeer_WithEnableRelay_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	adminInfo, err := db.NewNACInfo(ctx, "", false)
	require.NoError(t, err)
	db, err := db.NewDB(ctx, store, adminInfo, dac.NoDocumentACP, nil)
	require.NoError(t, err)
	defer db.Close()
	n, err := NewPeer(
		context.Background(),
		db.Events(),
		immutable.None[dac.DocumentACP](),
		db,
		config.WithEnableRelay(true),
	)
	require.NoError(t, err)
	n.Close()
}

func TestNewPeer_NoPubSub_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	adminInfo, err := db.NewNACInfo(ctx, "", false)
	require.NoError(t, err)
	db, err := db.NewDB(ctx, store, adminInfo, dac.NoDocumentACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n, err := NewPeer(
		context.Background(),
		db.Events(),
		immutable.None[dac.DocumentACP](),
		db,
		config.WithEnablePubSub(false),
	)
	require.NoError(t, err)
	require.Nil(t, n.ps)
	n.Close()
}

func TestNewPeer_WithEnablePubSub_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	adminInfo, err := db.NewNACInfo(ctx, "", false)
	require.NoError(t, err)
	db, err := db.NewDB(ctx, store, adminInfo, dac.NoDocumentACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n, err := NewPeer(
		ctx,
		db.Events(),
		immutable.None[dac.DocumentACP](),
		db,
		config.WithEnablePubSub(true),
	)

	require.NoError(t, err)
	// overly simple check of validity of pubsub, avoiding the process of creating a PubSub
	require.NotNil(t, n.ps)
	n.Close()
}

func TestNodeClose_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	adminInfo, err := db.NewNACInfo(ctx, "", false)
	require.NoError(t, err)
	db, err := db.NewDB(ctx, store, adminInfo, dac.NoDocumentACP, nil)
	require.NoError(t, err)
	defer db.Close()
	n, err := NewPeer(
		context.Background(),
		db.Events(),
		immutable.None[dac.DocumentACP](),
		db,
	)
	require.NoError(t, err)
	n.Close()
}

func TestListenAddrs_WithListenAddresses_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	adminInfo, err := db.NewNACInfo(ctx, "", false)
	require.NoError(t, err)
	db, err := db.NewDB(ctx, store, adminInfo, dac.NoDocumentACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n, err := NewPeer(
		context.Background(),
		db.Events(),
		immutable.None[dac.DocumentACP](),
		db,
		config.WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
	)
	require.NoError(t, err)
	require.Contains(t, n.ListenAddrs()[0].String(), "/tcp/")
	n.Close()
}

func TestPeer_WithBootstrapPeers_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	adminInfo, err := db.NewNACInfo(ctx, "", false)
	require.NoError(t, err)
	db, err := db.NewDB(ctx, store, adminInfo, dac.NoDocumentACP, nil)
	require.NoError(t, err)
	defer db.Close()

	n, err := NewPeer(
		context.Background(),
		db.Events(),
		immutable.None[dac.DocumentACP](),
		db,
		config.WithBootstrapPeers("/ip4/127.0.0.1/tcp/6666/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ"),
	)
	require.NoError(t, err)

	n.Close()
}
