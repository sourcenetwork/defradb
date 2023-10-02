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
	"sync"
	"testing"
	"time"

	dag "github.com/ipfs/boxo/ipld/merkledag"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	ipld "github.com/ipfs/go-ipld-format"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/core/crdt"
	netutils "github.com/sourcenetwork/defradb/net/utils"
)

const timeout = 5 * time.Second

func TestSendJobWorker_ExitOnContextClose_NoError(t *testing.T) {
	ctx := context.Background()
	_, n := newTestNode(ctx, t)
	done := make(chan struct{})
	go func() {
		n.sendJobWorker()
		close(done)
	}()
	err := n.Close()
	require.NoError(t, err)
	select {
	case <-done:
	case <-time.After(timeout):
		t.Error("failed to close sendJobWorker")
	}
}

func TestSendJobWorker_WithNewJobWithClosePriorToProcessing_NoError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	done := make(chan struct{})
	go func() {
		n.sendJobWorker()
		close(done)
	}()
	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`))
	require.NoError(t, err)
	dsKey := core.DataStoreKeyFromDocKey(doc.Key())

	txn, err := db.NewTxn(ctx, false)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	n.sendJobs <- &dagJob{
		session:    &wg,
		node:       &EmptyNode{},
		collection: col,
		dsKey:      dsKey,
		txn:        txn,
	}

	err = n.Close()
	require.NoError(t, err)
	select {
	case <-done:
	case <-time.After(timeout):
		t.Error("failed to close sendJobWorker")
	}
}

func TestSendJobWorker_WithNewJob_NoError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	done := make(chan struct{})
	go func() {
		n.sendJobWorker()
		close(done)
	}()
	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`))
	require.NoError(t, err)
	dsKey := core.DataStoreKeyFromDocKey(doc.Key())

	txn, err := db.NewTxn(ctx, false)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	n.sendJobs <- &dagJob{
		session:    &wg,
		node:       &EmptyNode{},
		collection: col,
		dsKey:      dsKey,
		txn:        txn,
	}
	// Give the jobworker time to process the job.
	time.Sleep(100 * time.Microsecond)
	err = n.Close()
	require.NoError(t, err)
	select {
	case <-done:
	case <-time.After(timeout):
		t.Error("failed to close sendJobWorker")
	}
}

func TestSendJobWorker_WithCloseJob_NoError(t *testing.T) {
	ctx := context.Background()
	db, n := newTestNode(ctx, t)
	done := make(chan struct{})
	go func() {
		n.sendJobWorker()
		close(done)
	}()
	_, err := db.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`))
	require.NoError(t, err)
	dsKey := core.DataStoreKeyFromDocKey(doc.Key())

	txn, err := db.NewTxn(ctx, false)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	n.sendJobs <- &dagJob{
		session:    &wg,
		node:       &EmptyNode{},
		collection: col,
		dsKey:      dsKey,
		txn:        txn,
	}

	n.closeJob <- dsKey.DocKey

	err = n.Close()
	require.NoError(t, err)
	select {
	case <-done:
	case <-time.After(timeout):
		t.Error("failed to close sendJobWorker")
	}
}

func TestSendJobWorker_WithPeerAndNoChildren_NoError(t *testing.T) {
	ctx := context.Background()
	db1, n1 := newTestNode(ctx, t)
	db2, n2 := newTestNode(ctx, t)

	addrs, err := netutils.ParsePeers([]string{n1.host.Addrs()[0].String() + "/p2p/" + n1.PeerID().String()})
	require.NoError(t, err)
	n2.Boostrap(addrs)

	done := make(chan struct{})
	go func() {
		n2.sendJobWorker()
		close(done)
	}()

	_, err = db1.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)
	_, err = db2.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db1.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`))
	require.NoError(t, err)
	dsKey := core.DataStoreKeyFromDocKey(doc.Key())

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	txn, err := db2.NewTxn(ctx, false)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	delta := &crdt.CompositeDAGDelta{
		SchemaVersionID: col.Schema().VersionID,
		Priority:        1,
		DocKey:          doc.Key().Bytes(),
	}

	node, err := makeNode(delta, []cid.Cid{})
	require.NoError(t, err)

	var getter format.NodeGetter = n2.Peer.newDAGSyncerTxn(txn)
	if sessionMaker, ok := getter.(SessionDAGSyncer); ok {
		log.Debug("Upgrading DAGSyncer with a session")
		getter = sessionMaker.Session(ctx)
	}

	n2.sendJobs <- &dagJob{
		session:    &wg,
		nodeGetter: getter,
		node:       node,
		collection: col,
		dsKey:      dsKey,
		txn:        txn,
	}
	// Give the jobworker time to process the job.
	time.Sleep(100 * time.Microsecond)
	err = n1.Close()
	require.NoError(t, err)
	err = n2.Close()
	require.NoError(t, err)
	select {
	case <-done:
	case <-time.After(timeout):
		t.Error("failed to close sendJobWorker")
	}
}

func TestSendJobWorker_WithPeerAndChildren_NoError(t *testing.T) {
	ctx := context.Background()
	db1, n1 := newTestNode(ctx, t)
	db2, n2 := newTestNode(ctx, t)

	addrs, err := netutils.ParsePeers([]string{n1.host.Addrs()[0].String() + "/p2p/" + n1.PeerID().String()})
	require.NoError(t, err)
	n2.Boostrap(addrs)

	done := make(chan struct{})
	go func() {
		n2.sendJobWorker()
		close(done)
	}()

	_, err = db1.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)
	_, err = db2.AddSchema(ctx, `type User {
		name: String
		age: Int
	}`)
	require.NoError(t, err)

	col, err := db1.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`))
	require.NoError(t, err)
	dsKey := core.DataStoreKeyFromDocKey(doc.Key())

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	txn, err := db2.NewTxn(ctx, false)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	links := []core.DAGLink{}
	for k := range doc.Fields() {
		delta := &crdt.LWWRegDelta{
			SchemaVersionID: col.Schema().VersionID,
			Priority:        1,
			DocKey:          doc.Key().Bytes(),
			FieldName:       k,
		}

		node, err := makeNode(delta, []cid.Cid{})
		require.NoError(t, err)

		links = append(links, core.DAGLink{
			Name: k,
			Cid:  node.Cid(),
		})
	}

	delta := &crdt.CompositeDAGDelta{
		SchemaVersionID: col.Schema().VersionID,
		Priority:        1,
		DocKey:          doc.Key().Bytes(),
		SubDAGs:         links,
	}

	node, err := makeNode(delta, []cid.Cid{})
	require.NoError(t, err)

	var getter format.NodeGetter = n2.Peer.newDAGSyncerTxn(txn)
	if sessionMaker, ok := getter.(SessionDAGSyncer); ok {
		log.Debug("Upgrading DAGSyncer with a session")
		getter = sessionMaker.Session(ctx)
	}

	n2.sendJobs <- &dagJob{
		session:    &wg,
		nodeGetter: getter,
		node:       node,
		collection: col,
		dsKey:      dsKey,
		txn:        txn,
	}
	// Give the jobworker time to process the job.
	time.Sleep(100 * time.Microsecond)
	err = n1.Close()
	require.NoError(t, err)
	err = n2.Close()
	require.NoError(t, err)
	select {
	case <-done:
	case <-time.After(timeout):
		t.Error("failed to close sendJobWorker")
	}
}

func makeNode(delta core.Delta, heads []cid.Cid) (ipld.Node, error) {
	var data []byte
	var err error
	if delta != nil {
		data, err = delta.Marshal()
		if err != nil {
			return nil, err
		}
	}

	nd := dag.NodeWithData(data)
	// The cid builder defaults to v0, we want to be using v1 Cids
	err = nd.SetCidBuilder(cid.V1Builder{
		Codec:    cid.DagProtobuf,
		MhType:   mh.SHA2_256,
		MhLength: -1,
	})
	if err != nil {
		return nil, err
	}

	// add heads
	for _, h := range heads {
		if err = nd.AddRawLink("_head", &ipld.Link{Cid: h}); err != nil {
			return nil, err
		}
	}

	// add delta specific links
	if comp, ok := delta.(core.CompositeDelta); ok {
		for _, dagLink := range comp.Links() {
			if err = nd.AddRawLink(dagLink.Name, &ipld.Link{Cid: dagLink.Cid}); err != nil {
				return nil, err
			}
		}
	}
	return nd, nil
}
