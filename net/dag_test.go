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

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/events"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/merkle/clock"
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
	n.Close()
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

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Definition())
	require.NoError(t, err)
	dsKey := core.DataStoreKeyFromDocID(doc.ID())

	txn, err := db.NewTxn(ctx, false)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	n.sendJobs <- &dagJob{
		session: &wg,
		bp: &blockProcessor{
			dsKey: dsKey,
			txn:   txn,
		},
	}
	// Give the jobworker time to process the job.
	time.Sleep(100 * time.Microsecond)
	n.Close()
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

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Definition())
	require.NoError(t, err)
	dsKey := core.DataStoreKeyFromDocID(doc.ID())

	txn, err := db.NewTxn(ctx, false)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	n.sendJobs <- &dagJob{
		session: &wg,
		bp: &blockProcessor{
			dsKey: dsKey,
			txn:   txn,
		},
	}

	n.closeJob <- dsKey.DocID

	n.Close()
	select {
	case <-done:
	case <-time.After(timeout):
		t.Error("failed to close sendJobWorker")
	}
}

func TestSendJobWorker_WithPeer_NoError(t *testing.T) {
	ctx := context.Background()
	db1, n1 := newTestNode(ctx, t)
	db2, n2 := newTestNode(ctx, t)

	sub1 := db1.Events().Subscribe(5, events.PeerEventName)
	defer db1.Events().Unsubscribe(sub1)

	sub2 := db2.Events().Subscribe(5, events.PeerEventName)
	defer db2.Events().Unsubscribe(sub2)

	addrs, err := netutils.ParsePeers([]string{n1.host.Addrs()[0].String() + "/p2p/" + n1.PeerID().String()})
	require.NoError(t, err)
	n2.Bootstrap(addrs)

	event1 := <-sub1.Value()
	assert.Equal(t, network.Connected, event1.(event.EvtPeerConnectednessChanged).Connectedness)
	assert.Equal(t, n2.PeerID(), event1.(event.EvtPeerConnectednessChanged).Peer)

	event2 := <-sub2.Value()
	assert.Equal(t, network.Connected, event2.(event.EvtPeerConnectednessChanged).Connectedness)
	assert.Equal(t, n1.PeerID(), event2.(event.EvtPeerConnectednessChanged).Peer)

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

	doc, err := client.NewDocFromJSON([]byte(`{"name": "John", "age": 30}`), col.Definition())
	require.NoError(t, err)
	dsKey := core.DataStoreKeyFromDocID(doc.ID())

	err = col.Create(ctx, doc)
	require.NoError(t, err)

	txn1, _ := db1.NewTxn(ctx, false)
	heads, _, err := clock.NewHeadSet(txn1.Headstore(), dsKey.ToHeadStoreKey().WithFieldId(core.COMPOSITE_NAMESPACE)).List(ctx)
	require.NoError(t, err)
	txn1.Discard(ctx)

	txn2, err := db2.NewTxn(ctx, false)
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	fetcher := n2.Peer.newDAGSyncerTxn(txn2)

	n2.sendJobs <- &dagJob{
		bp:      newBlockProcessor(n2.Peer, txn2, col, dsKey, fetcher),
		session: &wg,
		cid:     heads[0],
	}
	wg.Wait()

	err = txn2.Commit(ctx)
	require.NoError(t, err)

	b, err := n1.db.Blockstore().Get(ctx, heads[0])
	require.NoError(t, err)
	block, err := coreblock.GetFromBytes(b.RawData())
	require.NoError(t, err)

	for _, link := range block.Links {
		exists, err := n2.db.Blockstore().Has(ctx, link.Cid)
		require.NoError(t, err)
		require.True(t, exists)
	}

	n1.Close()
	n2.Close()
	select {
	case <-done:
	case <-time.After(timeout):
		t.Error("failed to close sendJobWorker")
	}
}
