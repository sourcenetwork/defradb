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
	ds "github.com/ipfs/go-datastore"
	ipld "github.com/ipfs/go-ipld-format"
	libp2p "github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	mh "github.com/multiformats/go-multihash"
	rpc "github.com/sourcenetwork/go-libp2p-pubsub-rpc"
	"github.com/stretchr/testify/require"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/datastore/memory"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/events"
	netutils "github.com/sourcenetwork/defradb/net/utils"
)

type EmptyNode struct{}

var ErrEmptyNode error = errors.New("dummy node")

func (n *EmptyNode) Resolve([]string) (any, []string, error) {
	return nil, nil, ErrEmptyNode
}

func (n *EmptyNode) Tree(string, int) []string {
	return nil
}

func (n *EmptyNode) ResolveLink([]string) (*ipld.Link, []string, error) {
	return nil, nil, ErrEmptyNode
}

func (n *EmptyNode) Copy() ipld.Node {
	return &EmptyNode{}
}

func (n *EmptyNode) Cid() cid.Cid {
	id, err := cid.V1Builder{
		Codec:    cid.DagProtobuf,
		MhType:   mh.SHA2_256,
		MhLength: 0, // default length
	}.Sum(nil)

	if err != nil {
		panic("failed to create an empty cid!")
	}
	return id
}

func (n *EmptyNode) Links() []*ipld.Link {
	return nil
}

func (n *EmptyNode) Loggable() map[string]any {
	return nil
}

func (n *EmptyNode) String() string {
	return "[]"
}

func (n *EmptyNode) RawData() []byte {
	return nil
}

func (n *EmptyNode) Size() (uint64, error) {
	return 0, nil
}

func (n *EmptyNode) Stat() (*ipld.NodeStat, error) {
	return &ipld.NodeStat{}, nil
}

func createCID(doc *client.Document) (cid.Cid, error) {
	pref := cid.V1Builder{
		Codec:    cid.DagProtobuf,
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

func newTestNode(ctx context.Context, t *testing.T) (client.DB, *Node) {
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, db.WithUpdateEvents())
	require.NoError(t, err)

	n, err := NewNode(
		ctx,
		db,
		WithListenAddresses(randomMultiaddr),
	)
	require.NoError(t, err)

	return db, n
}

func TestNewPeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, db.WithUpdateEvents())
	require.NoError(t, err)

	h, err := libp2p.New()
	require.NoError(t, err)

	_, err = NewPeer(ctx, db, h, nil, nil, nil, nil)
	require.NoError(t, err)
}

func TestNewPeer_NoDB_NilDBError(t *testing.T) {
	ctx := context.Background()

	h, err := libp2p.New()
	require.NoError(t, err)

	_, err = NewPeer(ctx, nil, h, nil, nil, nil, nil)
	require.ErrorIs(t, err, ErrNilDB)
}

func TestNewPeer_WithExistingTopic_TopicAlreadyExistsError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, db.WithUpdateEvents())
	require.NoError(t, err)

	_, err = db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Schema())
	require.NoError(t, err)

	err = col.Create(ctx, acpIdentity.NoIdentity, doc)
	require.NoError(t, err)

	h, err := libp2p.New()
	require.NoError(t, err)

	ps, err := pubsub.NewGossipSub(
		ctx,
		h,
		pubsub.WithPeerExchange(true),
		pubsub.WithFloodPublish(true),
	)
	require.NoError(t, err)

	_, err = rpc.NewTopic(ctx, ps, h.ID(), doc.ID().String(), true)
	require.NoError(t, err)

	_, err = NewPeer(ctx, db, h, nil, ps, nil, nil)
	require.ErrorContains(t, err, "topic already exists")
}

func TestStartAndClose_NoError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	err := n.Start()
	require.NoError(t, err)

	db.Close()
}

func TestStart_WithKnownPeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db1, err := db.NewDB(ctx, store, db.WithUpdateEvents())
	require.NoError(t, err)

	store2 := memory.NewDatastore(ctx)
	db2, err := db.NewDB(ctx, store2, db.WithUpdateEvents())
	require.NoError(t, err)

	n1, err := NewNode(
		ctx,
		db1,
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	n2, err := NewNode(
		ctx,
		db2,
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)

	addrs, err := netutils.ParsePeers([]string{n1.host.Addrs()[0].String() + "/p2p/" + n1.PeerID().String()})
	if err != nil {
		t.Fatal(err)
	}
	n2.Bootstrap(addrs)

	err = n2.Start()
	require.NoError(t, err)

	db1.Close()
	db2.Close()
}

func TestStart_WithOfflineKnownPeer_NoError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db1, err := db.NewDB(ctx, store, db.WithUpdateEvents())
	require.NoError(t, err)

	store2 := memory.NewDatastore(ctx)
	db2, err := db.NewDB(ctx, store2, db.WithUpdateEvents())
	require.NoError(t, err)

	n1, err := NewNode(
		ctx,
		db1,
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)
	n2, err := NewNode(
		ctx,
		db2,
		WithListenAddresses("/ip4/0.0.0.0/tcp/0"),
	)
	require.NoError(t, err)

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

	db1.Close()
	db2.Close()
}

func TestStart_WithNoUpdateChannel_NilUpdateChannelError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store)
	require.NoError(t, err)

	n, err := NewNode(
		ctx,
		db,
		WithEnablePubSub(true),
	)
	require.NoError(t, err)

	err = n.Start()
	require.ErrorIs(t, err, ErrNilUpdateChannel)

	db.Close()
}

func TestStart_WitClosedUpdateChannel_ClosedChannelError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewDatastore(ctx)
	db, err := db.NewDB(ctx, store, db.WithUpdateEvents())
	require.NoError(t, err)

	n, err := NewNode(
		ctx,
		db,
		WithEnablePubSub(true),
	)
	require.NoError(t, err)

	db.Events().Updates.Value().Close()

	err = n.Start()
	require.ErrorContains(t, err, "cannot subscribe to a closed channel")

	db.Close()
}

func TestRegisterNewDocument_NoError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Schema())
	require.NoError(t, err)

	cid, err := createCID(doc)
	require.NoError(t, err)

	err = n.RegisterNewDocument(ctx, doc.ID(), cid, &EmptyNode{}, col.SchemaRoot())
	require.NoError(t, err)
}

func TestRegisterNewDocument_RPCTopicAlreadyRegisteredError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Schema())
	require.NoError(t, err)

	_, err = rpc.NewTopic(ctx, n.Peer.ps, n.Peer.host.ID(), doc.ID().String(), true)
	require.NoError(t, err)

	cid, err := createCID(doc)
	require.NoError(t, err)

	err = n.RegisterNewDocument(ctx, doc.ID(), cid, &EmptyNode{}, col.SchemaRoot())
	require.Equal(t, err.Error(), "creating topic: joining topic: topic already exists")
}

func TestSetReplicator_NoError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	info, err := peer.AddrInfoFromString("/ip4/0.0.0.0/tcp/0/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	require.NoError(t, err)

	err = n.Peer.SetReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"User"},
	})
	require.NoError(t, err)
}

func TestSetReplicator_WithInvalidAddress_EmptyPeerIDError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	err = n.Peer.SetReplicator(ctx, client.Replicator{
		Info:    peer.AddrInfo{},
		Schemas: []string{"User"},
	})
	require.ErrorContains(t, err, "empty peer ID")
}

func TestSetReplicator_WithDBClosed_DatastoreClosedError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	db.Close()

	info, err := peer.AddrInfoFromString("/ip4/0.0.0.0/tcp/0/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	require.NoError(t, err)

	err = n.Peer.SetReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"User"},
	})
	require.ErrorContains(t, err, "datastore closed")
}

func TestSetReplicator_WithUndefinedCollection_KeyNotFoundError(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	defer n.Close()

	info, err := peer.AddrInfoFromString("/ip4/0.0.0.0/tcp/0/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	require.NoError(t, err)

	err = n.Peer.SetReplicator(ctx, client.Replicator{
		Info:    *info,
		Schemas: []string{"User"},
	})
	require.ErrorContains(t, err, "failed to get collections for replicator: datastore: key not found")
}

func TestSetReplicator_ForAllCollections_NoError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	info, err := peer.AddrInfoFromString("/ip4/0.0.0.0/tcp/0/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N")
	require.NoError(t, err)

	err = n.Peer.SetReplicator(ctx, client.Replicator{
		Info: *info,
	})
	require.NoError(t, err)
}

func TestPushToReplicator_SingleDocumentNoPeer_FailedToReplicateLogError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()
	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Schema())
	require.NoError(t, err)

	err = col.Create(ctx, acpIdentity.NoIdentity, doc)
	require.NoError(t, err)

	keysCh, err := col.GetAllDocIDs(ctx, acpIdentity.NoIdentity)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)

	n.pushToReplicator(ctx, txn, col, keysCh, n.PeerID())
}

func TestDeleteReplicator_WithDBClosed_DataStoreClosedError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	info := peer.AddrInfo{
		ID:    n.PeerID(),
		Addrs: n.ListenAddrs(),
	}

	db.Close()

	err := n.Peer.DeleteReplicator(ctx, client.Replicator{
		Info:    info,
		Schemas: []string{"User"},
	})
	require.ErrorContains(t, err, "datastore closed")
}

func TestDeleteReplicator_WithTargetSelf_SelfTargetForReplicatorError(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	defer n.Close()

	err := n.Peer.DeleteReplicator(ctx, client.Replicator{
		Info:    n.PeerInfo(),
		Schemas: []string{"User"},
	})
	require.ErrorIs(t, err, ErrSelfTargetForReplicator)
}

func TestDeleteReplicator_WithInvalidCollection_KeyNotFoundError(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	defer n.Close()

	_, n2 := newTestNode(ctx, t)
	defer n2.Close()

	err := n.Peer.DeleteReplicator(ctx, client.Replicator{
		Info:    n2.PeerInfo(),
		Schemas: []string{"User"},
	})
	require.ErrorContains(t, err, "failed to get collections for replicator: datastore: key not found")
}

func TestDeleteReplicator_WithCollectionAndPreviouslySetReplicator_NoError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	_, n2 := newTestNode(ctx, t)
	defer n2.Close()

	err = n.Peer.SetReplicator(ctx, client.Replicator{
		Info: n2.PeerInfo(),
	})
	require.NoError(t, err)

	err = n.Peer.DeleteReplicator(ctx, client.Replicator{
		Info: n2.PeerInfo(),
	})
	require.NoError(t, err)
}

func TestDeleteReplicator_WithNoCollection_NoError(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	defer n.Close()

	_, n2 := newTestNode(ctx, t)
	defer n2.Close()

	err := n.Peer.DeleteReplicator(ctx, client.Replicator{
		Info: n2.PeerInfo(),
	})
	require.NoError(t, err)
}

func TestDeleteReplicator_WithNotSetReplicator_KeyNotFoundError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	_, n2 := newTestNode(ctx, t)
	defer n2.Close()

	err = n.Peer.DeleteReplicator(ctx, client.Replicator{
		Info:    n2.PeerInfo(),
		Schemas: []string{"User"},
	})
	require.ErrorContains(t, err, "datastore: key not found")
}

func TestGetAllReplicator_WithReplicator_NoError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	_, n2 := newTestNode(ctx, t)
	defer n2.Close()

	err = n.Peer.SetReplicator(ctx, client.Replicator{
		Info: n2.PeerInfo(),
	})
	require.NoError(t, err)

	reps, err := n.Peer.GetAllReplicators(ctx)
	require.NoError(t, err)

	require.Len(t, reps, 1)
	require.Equal(t, n2.PeerInfo().ID, reps[0].Info.ID)
}

func TestGetAllReplicator_WithDBClosed_DatastoreClosedError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	db.Close()

	_, err := n.Peer.GetAllReplicators(ctx)
	require.ErrorContains(t, err, "datastore closed")
}

func TestLoadReplicators_WithDBClosed_DatastoreClosedError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	db.Close()

	err := n.Peer.loadReplicators(ctx)
	require.ErrorContains(t, err, "datastore closed")
}

func TestLoadReplicator_WithReplicator_NoError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	_, n2 := newTestNode(ctx, t)
	defer n2.Close()

	err = n.Peer.SetReplicator(ctx, client.Replicator{
		Info: n2.PeerInfo(),
	})
	require.NoError(t, err)

	err = n.Peer.loadReplicators(ctx)
	require.NoError(t, err)
}

func TestLoadReplicator_WithReplicatorAndEmptyReplicatorMap_NoError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	_, n2 := newTestNode(ctx, t)
	defer n2.Close()

	err = n.Peer.SetReplicator(ctx, client.Replicator{
		Info: n2.PeerInfo(),
	})
	require.NoError(t, err)

	n.replicators = make(map[string]map[peer.ID]struct{})

	err = n.Peer.loadReplicators(ctx)
	require.NoError(t, err)
}

func TestAddP2PCollections_WithInvalidCollectionID_NotFoundError(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	defer n.Close()

	err := n.Peer.AddP2PCollections(ctx, []string{"invalid_collection"})
	require.Error(t, err, ds.ErrNotFound)
}

func TestAddP2PCollections_NoError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	err = n.Peer.AddP2PCollections(ctx, []string{col.SchemaRoot()})
	require.NoError(t, err)
}

func TestRemoveP2PCollectionsWithInvalidCollectionID(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	defer n.Close()

	err := n.Peer.RemoveP2PCollections(ctx, []string{"invalid_collection"})
	require.Error(t, err, ds.ErrNotFound)
}

func TestRemoveP2PCollections(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	err = n.Peer.RemoveP2PCollections(ctx, []string{col.SchemaRoot()})
	require.NoError(t, err)
}

func TestGetAllP2PCollectionsWithNoCollections(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	defer n.Close()

	cols, err := n.Peer.GetAllP2PCollections(ctx)
	require.NoError(t, err)
	require.Len(t, cols, 0)
}

func TestGetAllP2PCollections(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	err = n.Peer.AddP2PCollections(ctx, []string{col.SchemaRoot()})
	require.NoError(t, err)

	cols, err := n.Peer.GetAllP2PCollections(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{col.SchemaRoot()}, cols)
}

func TestHandleDocCreateLog_NoError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Schema())
	require.NoError(t, err)

	err = col.Create(ctx, acpIdentity.NoIdentity, doc)
	require.NoError(t, err)

	docCid, err := createCID(doc)
	require.NoError(t, err)

	delta := &crdt.CompositeDAGDelta{
		SchemaVersionID: col.Schema().VersionID,
		Priority:        1,
		DocID:           doc.ID().Bytes(),
	}

	node, err := makeNode(delta, []cid.Cid{docCid})
	require.NoError(t, err)

	err = n.handleDocCreateLog(events.Update{
		DocID:      doc.ID().String(),
		Cid:        docCid,
		SchemaRoot: col.SchemaRoot(),
		Block:      node,
		Priority:   0,
	})
	require.NoError(t, err)
}

func TestHandleDocCreateLog_WithInvalidDocID_NoError(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	defer n.Close()

	err := n.handleDocCreateLog(events.Update{
		DocID: "some-invalid-key",
	})
	require.ErrorContains(t, err, "failed to get DocID from broadcast message: selected encoding not supported")
}

func TestHandleDocCreateLog_WithExistingTopic_TopicExistsError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Schema())
	require.NoError(t, err)

	err = col.Create(ctx, acpIdentity.NoIdentity, doc)
	require.NoError(t, err)

	_, err = rpc.NewTopic(ctx, n.ps, n.host.ID(), doc.ID().String(), true)
	require.NoError(t, err)

	err = n.handleDocCreateLog(events.Update{
		DocID:      doc.ID().String(),
		SchemaRoot: col.SchemaRoot(),
	})
	require.ErrorContains(t, err, "topic already exists")
}

func TestHandleDocUpdateLog_NoError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Schema())
	require.NoError(t, err)

	err = col.Create(ctx, acpIdentity.NoIdentity, doc)
	require.NoError(t, err)

	docCid, err := createCID(doc)
	require.NoError(t, err)

	delta := &crdt.CompositeDAGDelta{
		SchemaVersionID: col.Schema().VersionID,
		Priority:        1,
		DocID:           doc.ID().Bytes(),
	}

	node, err := makeNode(delta, []cid.Cid{docCid})
	require.NoError(t, err)

	err = n.handleDocUpdateLog(events.Update{
		DocID:      doc.ID().String(),
		Cid:        docCid,
		SchemaRoot: col.SchemaRoot(),
		Block:      node,
		Priority:   0,
	})
	require.NoError(t, err)
}

func TestHandleDoUpdateLog_WithInvalidDocID_NoError(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	defer n.Close()

	err := n.handleDocUpdateLog(events.Update{
		DocID: "some-invalid-key",
	})
	require.ErrorContains(t, err, "failed to get DocID from broadcast message: selected encoding not supported")
}

func TestHandleDocUpdateLog_WithExistingDocIDTopic_TopicExistsError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Schema())
	require.NoError(t, err)

	err = col.Create(ctx, acpIdentity.NoIdentity, doc)
	require.NoError(t, err)

	docCid, err := createCID(doc)
	require.NoError(t, err)

	delta := &crdt.CompositeDAGDelta{
		SchemaVersionID: col.Schema().VersionID,
		Priority:        1,
		DocID:           doc.ID().Bytes(),
	}

	node, err := makeNode(delta, []cid.Cid{docCid})
	require.NoError(t, err)

	_, err = rpc.NewTopic(ctx, n.ps, n.host.ID(), doc.ID().String(), true)
	require.NoError(t, err)

	err = n.handleDocUpdateLog(events.Update{
		DocID:      doc.ID().String(),
		Cid:        docCid,
		SchemaRoot: col.SchemaRoot(),
		Block:      node,
	})
	require.ErrorContains(t, err, "topic already exists")
}

func TestHandleDocUpdateLog_WithExistingSchemaTopic_TopicExistsError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	defer n.Close()

	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Schema())
	require.NoError(t, err)

	err = col.Create(ctx, acpIdentity.NoIdentity, doc)
	require.NoError(t, err)

	docCid, err := createCID(doc)
	require.NoError(t, err)

	delta := &crdt.CompositeDAGDelta{
		SchemaVersionID: col.Schema().VersionID,
		Priority:        1,
		DocID:           doc.ID().Bytes(),
	}

	node, err := makeNode(delta, []cid.Cid{docCid})
	require.NoError(t, err)

	_, err = rpc.NewTopic(ctx, n.ps, n.host.ID(), col.SchemaRoot(), true)
	require.NoError(t, err)

	err = n.handleDocUpdateLog(events.Update{
		DocID:      doc.ID().String(),
		Cid:        docCid,
		SchemaRoot: col.SchemaRoot(),
		Block:      node,
	})
	require.ErrorContains(t, err, "topic already exists")
}

func TestSession_NoError(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	defer n.Close()
	ng := n.Session(ctx)
	require.Implements(t, (*ipld.NodeGetter)(nil), ng)
}
