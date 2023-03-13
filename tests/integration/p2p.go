// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
	netutils "github.com/sourcenetwork/defradb/net/utils"
	"github.com/sourcenetwork/defradb/node"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ConnectPeers connects two nodes together as peers.
//
// Updates between shared documents should be synced in either direction,
// but new documents will only be synced if explicitly requested (e.g. via
// collection subscription).
type ConnectPeers struct {
	// SourceNodeID is the node ID (index) of the first node to connect.
	//
	// Is completely interchangable with TargetNodeID and which way round
	// these properties are specified is purely cosmetic.
	SourceNodeID int

	// TargetNodeID is the node ID (index) of the second node to connect.
	//
	// Is completely interchangable with SourceNodeID and which way round
	// these properties are specified is purely cosmetic.
	TargetNodeID int
}

// ConfigureReplicator confugures a directional replicator relationship between
// two nodes.
//
// All document changes made in the source node will be synced to the target node.
// New documents created in the target node will not be synced to the source node,
// however updates in the target node to documents synced from the source node will
// be synced back to the source node.
type ConfigureReplicator struct {
	// SourceNodeID is the node ID (index) of the node from which data should be replicated.
	SourceNodeID int

	// TargetNodeID is the node ID (index) of the node to which data should be replicated.
	TargetNodeID int
}

// SubscribeToCollection sets up a subscription on the given node to the given collection.
//
// Changes made to subscribed collections in peers connected to this node will be synced from
// them to this node.
type SubscribeToCollection struct {
	// NodeID is the node ID (index) of the node in which to activate the subscription.
	//
	// Changes made to subscribed collections in peers connected to this node will be synced from
	// them to this node.
	NodeID int

	// CollectionID is the collection ID (index) of the collection to subscribe to.
	CollectionID int
}

// WaitForSync is an action that instructs the test framework to wait for all document synchronization
// to complete before progressing.
//
// For example you will likely wish to `WaitForSync` after creating a document in node 0 before querying
// node 1 to see if it has been replicated.
type WaitForSync struct{}

// AnyOf may be used as `Results` field where the value may
// be one of several values, yet the value of that field must be the same
// across all nodes due to strong eventual consistancy.
type AnyOf []any

// connectPeers connects two existing, started, nodes as peers.  It returns a channel
// that will recieve an empty struct upon sync completion of all expected peer-sync events.
//
// Any errors generated whilst configuring the peers or waiting on sync will result in a test failure.
func connectPeers(
	ctx context.Context,
	t *testing.T,
	testCase TestCase,
	cfg ConnectPeers,
	nodes []*node.Node,
	addresses []string,
) chan struct{} {
	sourceNode := nodes[cfg.SourceNodeID]
	targetNode := nodes[cfg.TargetNodeID]
	targetAddress := addresses[cfg.TargetNodeID]

	log.Info(ctx, "Parsing bootstrap peers", logging.NewKV("Peers", targetAddress))
	addrs, err := netutils.ParsePeers([]string{targetAddress})
	if err != nil {
		t.Fatal(fmt.Sprintf("failed to parse bootstrap peers %v", targetAddress), err)
	}
	log.Info(ctx, "Bootstrapping with peers", logging.NewKV("Addresses", addrs))
	sourceNode.Boostrap(addrs)

	// Boostrap triggers a bunch of async stuff for which we have no good way of waiting on.  It must be
	// allowed to complete before documentation begins or it will not even try and sync it. So for now, we
	// sleep a little.
	time.Sleep(100 * time.Millisecond)

	nodeCollections := map[int][]int{}
	for _, a := range testCase.Actions {
		switch action := a.(type) {
		case SubscribeToCollection:
			// Node collections must be populated before re-iterating through the full action set as
			// documents created before the subscription must still be waited on.
			nodeCollections[action.NodeID] = append(nodeCollections[action.NodeID], action.CollectionID)
		}
	}

	sourceToTargetEvents := []int{0}
	targetToSourceEvents := []int{0}
	waitIndex := 0
	for _, a := range testCase.Actions {
		switch action := a.(type) {
		case CreateDoc:
			sourceCollectionSubscribed := collectionSubscribedTo(nodeCollections, cfg.SourceNodeID, action.CollectionID)
			targetCollectionSubscribed := collectionSubscribedTo(nodeCollections, cfg.TargetNodeID, action.CollectionID)

			// Peers sync trigger sync events for documents that exist prior to configuration, even if they already
			// exist at the destination, so we need to wait for documents created on all nodes, as well as those
			// created on the target.
			if (!action.NodeID.HasValue() ||
				action.NodeID.Value() == cfg.TargetNodeID) &&
				targetCollectionSubscribed {
				sourceToTargetEvents[waitIndex] += 1
			}

			// Peers sync trigger sync events for documents that exist prior to configuration, even if they already
			// exist at the destination, so we need to wait for documents created on all nodes, as well as those
			// created on the source.
			if (!action.NodeID.HasValue() ||
				action.NodeID.Value() == cfg.SourceNodeID) &&
				sourceCollectionSubscribed {
				targetToSourceEvents[waitIndex] += 1
			}

		case UpdateDoc:
			// Updates to existing docs should always sync (no-sub required)
			if action.NodeID.HasValue() && action.NodeID.Value() == cfg.TargetNodeID {
				targetToSourceEvents[waitIndex] += 1
			}
			if action.NodeID.HasValue() && action.NodeID.Value() == cfg.SourceNodeID {
				sourceToTargetEvents[waitIndex] += 1
			}

		case WaitForSync:
			waitIndex += 1
			targetToSourceEvents = append(targetToSourceEvents, 0)
			sourceToTargetEvents = append(sourceToTargetEvents, 0)
		}
	}

	nodeSynced := make(chan struct{})
	go func() {
		for waitIndex := 0; waitIndex < len(sourceToTargetEvents); waitIndex++ {
			for i := 0; i < targetToSourceEvents[waitIndex]; i++ {
				err := sourceNode.WaitForPushLogByPeerEvent(targetNode.PeerID())
				require.NoError(t, err)
			}
			for i := 0; i < sourceToTargetEvents[waitIndex]; i++ {
				err := targetNode.WaitForPushLogByPeerEvent(sourceNode.PeerID())
				require.NoError(t, err)
			}
			nodeSynced <- struct{}{}
		}
	}()

	return nodeSynced
}

// collectionSubscribedTo returns true if the collection on the given node
// has been subscribed to.
func collectionSubscribedTo(
	nodeCollections map[int][]int,
	nodeID int,
	collectionID int,
) bool {
	targetSubscriptionCollections := nodeCollections[nodeID]
	for _, collectionId := range targetSubscriptionCollections {
		if collectionId == collectionID {
			return true
		}
	}
	return false
}

// configureReplicator configures a replicator relationship between two existing, staarted, nodes.
// It returns a channel that will recieve an empty struct upon sync completion of all expected
// replicator-sync events.
//
// Any errors generated whilst configuring the peers or waiting on sync will result in a test failure.
func configureReplicator(
	ctx context.Context,
	t *testing.T,
	testCase TestCase,
	cfg ConfigureReplicator,
	nodes []*node.Node,
	addresses []string,
) chan struct{} {
	sourceNode := nodes[cfg.SourceNodeID]
	targetNode := nodes[cfg.TargetNodeID]
	targetAddress := addresses[cfg.TargetNodeID]

	addr, err := ma.NewMultiaddr(targetAddress)
	require.NoError(t, err)

	_, err = sourceNode.Peer.SetReplicator(ctx, addr)
	require.NoError(t, err)

	sourceToTargetEvents := []int{0}
	targetToSourceEvents := []int{0}
	waitIndex := 0
	for _, a := range testCase.Actions {
		switch action := a.(type) {
		case CreateDoc:
			if action.NodeID.HasValue() && action.NodeID.Value() == cfg.SourceNodeID {
				sourceToTargetEvents[waitIndex] += 1
			}

		case UpdateDoc:
			if action.NodeID.HasValue() && action.NodeID.Value() == cfg.TargetNodeID {
				targetToSourceEvents[waitIndex] += 1
			}
			if action.NodeID.HasValue() && action.NodeID.Value() == cfg.SourceNodeID {
				sourceToTargetEvents[waitIndex] += 1
			}

		case WaitForSync:
			waitIndex += 1
			targetToSourceEvents = append(targetToSourceEvents, 0)
			sourceToTargetEvents = append(sourceToTargetEvents, 0)
		}
	}

	nodeSynced := make(chan struct{})
	go func() {
		for waitIndex := 0; waitIndex < len(sourceToTargetEvents); waitIndex++ {
			for i := 0; i < targetToSourceEvents[waitIndex]; i++ {
				err := sourceNode.WaitForPushLogByPeerEvent(targetNode.PeerID())
				require.NoError(t, err)
			}
			for i := 0; i < sourceToTargetEvents[waitIndex]; i++ {
				err := targetNode.WaitForPushLogByPeerEvent(sourceNode.PeerID())
				require.NoError(t, err)
			}
			nodeSynced <- struct{}{}
		}
	}()

	return nodeSynced
}

// subscribeToCollection sets up a collection subscription on the given node/collection.
//
// Any errors generated during this process will result in a test failure.
func subscribeToCollection(
	ctx context.Context,
	t *testing.T,
	action SubscribeToCollection,
	nodes []*node.Node,
	collections [][]client.Collection,
) {
	n := nodes[action.NodeID]
	col := collections[action.NodeID][action.CollectionID]

	err := n.Peer.AddP2PCollections([]string{col.SchemaID()})
	require.NoError(t, err)

	// The `n.Peer.AddP2PCollections(colIDs)` call above is calling some asynchronous functions
	// for the pubsub subscription and those functions can take a bit of time to complete,
	// we need to make sure this has finished before progressing.
	time.Sleep(100 * time.Millisecond)
}

// waitForSync waits for all given wait channels to receive an item signaling completion.
//
// Will fail the test if an event is not received within the expected time interval to prevent tests
// from running forever.
func waitForSync(
	t *testing.T,
	testCase TestCase,
	action WaitForSync,
	waitChans []chan struct{},
) {
	for _, resultsChan := range waitChans {
		select {
		case <-resultsChan:
			continue

		// a safety in case the stream hangs - we don't want the tests to run forever.
		case <-time.After(subscriptionTimeout * 10):
			assert.Fail(t, "timeout occured while waiting for data stream", testCase.Description)
		}
	}
}

const randomMultiaddr = "/ip4/0.0.0.0/tcp/0"

func RandomNetworkingConfig() ConfigureNode {
	cfg := config.DefaultConfig()
	cfg.Net.P2PAddress = randomMultiaddr
	cfg.Net.RPCAddress = "0.0.0.0:0"
	cfg.Net.TCPAddress = randomMultiaddr

	return ConfigureNode{
		Config: *cfg,
	}
}
