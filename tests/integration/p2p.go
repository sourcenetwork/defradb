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
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/tests/clients"

	"github.com/libp2p/go-libp2p/core/peer"
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
	// Is completely interchangeable with TargetNodeID and which way round
	// these properties are specified is purely cosmetic.
	SourceNodeID int

	// TargetNodeID is the node ID (index) of the second node to connect.
	//
	// Is completely interchangeable with SourceNodeID and which way round
	// these properties are specified is purely cosmetic.
	TargetNodeID int
}

// ConfigureReplicator configures a directional replicator relationship between
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

// DeleteReplicator deletes a directional replicator relationship between two nodes.
type DeleteReplicator struct {
	// SourceNodeID is the node ID (index) of the node from which the replicator should be deleted.
	SourceNodeID int

	// TargetNodeID is the node ID (index) of the node to which the replicator should be deleted.
	TargetNodeID int
}

const (
	// NonExistentCollectionID can be used to represent a non-existent collection ID, it will be substituted
	// for a non-existent collection ID when used in actions that support this.
	NonExistentCollectionID       int    = -1
	NonExistentCollectionSchemaID string = "NonExistentCollectionID"
)

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

	// CollectionIDs are the collection IDs (indexes) of the collections to subscribe to.
	//
	// A [NonExistentCollectionID] may be provided to test non-existent collection IDs.
	CollectionIDs []int

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// UnsubscribeToCollection removes the given collections from the set of active subscriptions on
// the given node.
type UnsubscribeToCollection struct {
	// NodeID is the node ID (index) of the node in which to remove the subscription.
	NodeID int

	// CollectionIDs are the collection IDs (indexes) of the collections to unsubscribe from.
	//
	// A [NonExistentCollectionID] may be provided to test non-existent collection IDs.
	CollectionIDs []int

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// GetAllP2PCollections gets the active subscriptions for the given node and compares them against the
// expected results.
type GetAllP2PCollections struct {
	// NodeID is the node ID (index) of the node in which to get the subscriptions for.
	NodeID int

	// ExpectedCollectionIDs are the collection IDs (indexes) of the collections expected.
	ExpectedCollectionIDs []int
}

// WaitForSync is an action that instructs the test framework to wait for all document synchronization
// to complete before progressing.
//
// For example you will likely wish to `WaitForSync` after creating a document in node 0 before querying
// node 1 to see if it has been replicated.
type WaitForSync struct {
	// ExpectedTimeout is the duration to wait when expecting a timeout to occur.
	ExpectedTimeout time.Duration
}

// connectPeers connects two existing, started, nodes as peers.  It returns a channel
// that will receive an empty struct upon sync completion of all expected peer-sync events.
//
// Any errors generated whilst configuring the peers or waiting on sync will result in a test failure.
func connectPeers(
	s *state,
	cfg ConnectPeers,
) {
	// If we have some database actions prior to connecting the peers, we want to ensure that they had time to
	// complete before we connect. Otherwise we might wrongly catch them in our wait function.
	time.Sleep(100 * time.Millisecond)
	sourceNode := s.nodes[cfg.SourceNodeID]
	targetNode := s.nodes[cfg.TargetNodeID]

	addrs := []peer.AddrInfo{targetNode.PeerInfo()}
	log.Info(s.ctx, "Bootstrapping with peers", logging.NewKV("Addresses", addrs))
	sourceNode.Bootstrap(addrs)

	// Bootstrap triggers a bunch of async stuff for which we have no good way of waiting on.  It must be
	// allowed to complete before documentation begins or it will not even try and sync it. So for now, we
	// sleep a little.
	time.Sleep(100 * time.Millisecond)
	setupPeerWaitSync(s, 0, cfg, sourceNode, targetNode)
}

func setupPeerWaitSync(
	s *state,
	startIndex int,
	cfg ConnectPeers,
	sourceNode clients.Client,
	targetNode clients.Client,
) {
	sourceToTargetEvents := []int{0}
	targetToSourceEvents := []int{0}

	sourcePeerInfo := sourceNode.PeerInfo()
	targetPeerInfo := targetNode.PeerInfo()

	nodeCollections := map[int][]int{}
	waitIndex := 0
	for i := startIndex; i < len(s.testCase.Actions); i++ {
		switch action := s.testCase.Actions[i].(type) {
		case SubscribeToCollection:
			if action.ExpectedError != "" {
				// If the subscription action is expected to error, then we should do nothing here.
				continue
			}
			// This is order dependent, items should be added in the same action-loop that reads them
			// as 'stuff' done before collection subscription should not be synced.
			nodeCollections[action.NodeID] = append(nodeCollections[action.NodeID], action.CollectionIDs...)

		case UnsubscribeToCollection:
			if action.ExpectedError != "" {
				// If the unsubscribe action is expected to error, then we should do nothing here.
				continue
			}

			// This is order dependent, items should be added in the same action-loop that reads them
			// as 'stuff' done before collection subscription should not be synced.
			existingCollectionIndexes := nodeCollections[action.NodeID]
			for _, collectionIndex := range action.CollectionIDs {
				for i, existingCollectionIndex := range existingCollectionIndexes {
					if collectionIndex == existingCollectionIndex {
						// Remove the matching collection index from the set:
						existingCollectionIndexes = append(existingCollectionIndexes[:i], existingCollectionIndexes[i+1:]...)
					}
				}
			}
			nodeCollections[action.NodeID] = existingCollectionIndexes

		case CreateDoc:
			sourceCollectionSubscribed := collectionSubscribedTo(nodeCollections, cfg.SourceNodeID, action.CollectionID)
			targetCollectionSubscribed := collectionSubscribedTo(nodeCollections, cfg.TargetNodeID, action.CollectionID)

			// Peers sync trigger sync events for documents that exist prior to configuration, even if they already
			// exist at the destination, so we need to wait for documents created on all nodes, as well as those
			// created on the target.
			if (!action.NodeID.HasValue() ||
				action.NodeID.Value() == cfg.TargetNodeID) &&
				sourceCollectionSubscribed {
				targetToSourceEvents[waitIndex] += 1
			}

			// Peers sync trigger sync events for documents that exist prior to configuration, even if they already
			// exist at the destination, so we need to wait for documents created on all nodes, as well as those
			// created on the source.
			if (!action.NodeID.HasValue() ||
				action.NodeID.Value() == cfg.SourceNodeID) &&
				targetCollectionSubscribed {
				sourceToTargetEvents[waitIndex] += 1
			}

		case DeleteDoc:
			// Updates to existing docs should always sync (no-sub required)
			if !action.DontSync && action.NodeID.HasValue() && action.NodeID.Value() == cfg.TargetNodeID {
				targetToSourceEvents[waitIndex] += 1
			}
			if !action.DontSync && action.NodeID.HasValue() && action.NodeID.Value() == cfg.SourceNodeID {
				sourceToTargetEvents[waitIndex] += 1
			}

		case UpdateDoc:
			// Updates to existing docs should always sync (no-sub required)
			if !action.DontSync && action.NodeID.HasValue() && action.NodeID.Value() == cfg.TargetNodeID {
				targetToSourceEvents[waitIndex] += 1
			}
			if !action.DontSync && action.NodeID.HasValue() && action.NodeID.Value() == cfg.SourceNodeID {
				sourceToTargetEvents[waitIndex] += 1
			}

		case WaitForSync:
			waitIndex += 1
			targetToSourceEvents = append(targetToSourceEvents, 0)
			sourceToTargetEvents = append(sourceToTargetEvents, 0)
		}
	}

	nodeSynced := make(chan struct{})
	ready := make(chan struct{})
	go func(ready chan struct{}) {
		ready <- struct{}{}
		for waitIndex := 0; waitIndex < len(sourceToTargetEvents); waitIndex++ {
			for i := 0; i < targetToSourceEvents[waitIndex]; i++ {
				err := sourceNode.WaitForPushLogByPeerEvent(targetPeerInfo.ID)
				require.NoError(s.t, err)
			}
			for i := 0; i < sourceToTargetEvents[waitIndex]; i++ {
				err := targetNode.WaitForPushLogByPeerEvent(sourcePeerInfo.ID)
				require.NoError(s.t, err)
			}
			nodeSynced <- struct{}{}
		}
	}(ready)
	// Ensure that the wait routine is ready to receive events before we continue.
	<-ready

	s.syncChans = append(s.syncChans, nodeSynced)
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

// configureReplicator configures a replicator relationship between two existing, started, nodes.
// It returns a channel that will receive an empty struct upon sync completion of all expected
// replicator-sync events.
//
// Any errors generated whilst configuring the peers or waiting on sync will result in a test failure.
func configureReplicator(
	s *state,
	cfg ConfigureReplicator,
) {
	// If we have some database actions prior to configuring the replicator, we want to ensure that they had time to
	// complete before the configuration. Otherwise we might wrongly catch them in our wait function.
	time.Sleep(100 * time.Millisecond)
	sourceNode := s.nodes[cfg.SourceNodeID]
	targetNode := s.nodes[cfg.TargetNodeID]

	err := sourceNode.SetReplicator(s.ctx, client.Replicator{
		Info: targetNode.PeerInfo(),
	})
	require.NoError(s.t, err)
	setupReplicatorWaitSync(s, 0, cfg, sourceNode, targetNode)
}

func deleteReplicator(
	s *state,
	cfg DeleteReplicator,
) {
	sourceNode := s.nodes[cfg.SourceNodeID]
	targetNode := s.nodes[cfg.TargetNodeID]

	err := sourceNode.DeleteReplicator(s.ctx, client.Replicator{
		Info: targetNode.PeerInfo(),
	})
	require.NoError(s.t, err)
}

func setupReplicatorWaitSync(
	s *state,
	startIndex int,
	cfg ConfigureReplicator,
	sourceNode clients.Client,
	targetNode clients.Client,
) {
	sourceToTargetEvents := []int{0}
	targetToSourceEvents := []int{0}

	sourcePeerInfo := sourceNode.PeerInfo()
	targetPeerInfo := targetNode.PeerInfo()

	docIDsSyncedToSource := map[int]struct{}{}
	waitIndex := 0
	currentDocID := 0
	for i := startIndex; i < len(s.testCase.Actions); i++ {
		switch action := s.testCase.Actions[i].(type) {
		case CreateDoc:
			if !action.NodeID.HasValue() || action.NodeID.Value() == cfg.SourceNodeID {
				docIDsSyncedToSource[currentDocID] = struct{}{}
			}

			// A document created on the source or one that is created on all nodes will be sent to the target even
			// it already has it. It will create a `received push log` event on the target which we need to wait for.
			if !action.NodeID.HasValue() || action.NodeID.Value() == cfg.SourceNodeID {
				sourceToTargetEvents[waitIndex] += 1
			}

			currentDocID++

		case DeleteDoc:
			if _, shouldSyncFromTarget := docIDsSyncedToSource[action.DocID]; shouldSyncFromTarget &&
				action.NodeID.HasValue() && action.NodeID.Value() == cfg.TargetNodeID {
				targetToSourceEvents[waitIndex] += 1
			}

			if action.NodeID.HasValue() && action.NodeID.Value() == cfg.SourceNodeID {
				sourceToTargetEvents[waitIndex] += 1
			}

		case UpdateDoc:
			if _, shouldSyncFromTarget := docIDsSyncedToSource[action.DocID]; shouldSyncFromTarget &&
				action.NodeID.HasValue() && action.NodeID.Value() == cfg.TargetNodeID {
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
	ready := make(chan struct{})
	go func(ready chan struct{}) {
		ready <- struct{}{}
		for waitIndex := 0; waitIndex < len(sourceToTargetEvents); waitIndex++ {
			for i := 0; i < targetToSourceEvents[waitIndex]; i++ {
				err := sourceNode.WaitForPushLogByPeerEvent(targetPeerInfo.ID)
				require.NoError(s.t, err)
			}
			for i := 0; i < sourceToTargetEvents[waitIndex]; i++ {
				err := targetNode.WaitForPushLogByPeerEvent(sourcePeerInfo.ID)
				require.NoError(s.t, err)
			}
			nodeSynced <- struct{}{}
		}
	}(ready)
	// Ensure that the wait routine is ready to receive events before we continue.
	<-ready

	s.syncChans = append(s.syncChans, nodeSynced)
}

// subscribeToCollection sets up a collection subscription on the given node/collection.
//
// Any errors generated during this process will result in a test failure.
func subscribeToCollection(
	s *state,
	action SubscribeToCollection,
) {
	n := s.nodes[action.NodeID]

	schemaIDs := []string{}
	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			schemaIDs = append(schemaIDs, NonExistentCollectionSchemaID)
			continue
		}

		col := s.collections[action.NodeID][collectionIndex]
		schemaIDs = append(schemaIDs, col.SchemaID())
	}

	err := n.AddP2PCollections(s.ctx, schemaIDs)
	expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

	// The `n.Peer.AddP2PCollections(colIDs)` call above is calling some asynchronous functions
	// for the pubsub subscription and those functions can take a bit of time to complete,
	// we need to make sure this has finished before progressing.
	time.Sleep(100 * time.Millisecond)
}

// unsubscribeToCollection removes the given collections from subscriptions on the given nodes.
//
// Any errors generated during this process will result in a test failure.
func unsubscribeToCollection(
	s *state,
	action UnsubscribeToCollection,
) {
	n := s.nodes[action.NodeID]

	schemaIDs := []string{}
	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			schemaIDs = append(schemaIDs, NonExistentCollectionSchemaID)
			continue
		}

		col := s.collections[action.NodeID][collectionIndex]
		schemaIDs = append(schemaIDs, col.SchemaID())
	}

	err := n.RemoveP2PCollections(s.ctx, schemaIDs)
	expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

	// The `n.Peer.RemoveP2PCollections(colIDs)` call above is calling some asynchronous functions
	// for the pubsub subscription and those functions can take a bit of time to complete,
	// we need to make sure this has finished before progressing.
	time.Sleep(100 * time.Millisecond)
}

// getAllP2PCollections gets all the active peer subscriptions and compares them against the
// given expected results.
//
// Any errors generated during this process will result in a test failure.
func getAllP2PCollections(
	s *state,
	action GetAllP2PCollections,
) {
	expectedCollections := []string{}
	for _, collectionIndex := range action.ExpectedCollectionIDs {
		col := s.collections[action.NodeID][collectionIndex]
		expectedCollections = append(expectedCollections, col.SchemaID())
	}

	n := s.nodes[action.NodeID]
	cols, err := n.GetAllP2PCollections(s.ctx)
	require.NoError(s.t, err)

	assert.Equal(s.t, expectedCollections, cols)
}

// waitForSync waits for all given wait channels to receive an item signaling completion.
//
// Will fail the test if an event is not received within the expected time interval to prevent tests
// from running forever.
func waitForSync(
	s *state,
	action WaitForSync,
) {
	var timeout time.Duration
	if action.ExpectedTimeout != 0 {
		timeout = action.ExpectedTimeout
	} else {
		timeout = subscriptionTimeout * 10
	}

	for _, resultsChan := range s.syncChans {
		select {
		case <-resultsChan:
			assert.True(
				s.t,
				action.ExpectedTimeout == 0,
				"unexpected document has been synced",
				s.testCase.Description,
			)

		// a safety in case the stream hangs - we don't want the tests to run forever.
		case <-time.After(timeout):
			assert.True(
				s.t,
				action.ExpectedTimeout != 0,
				"timeout occurred while waiting for data stream",
				s.testCase.Description,
			)
		}
	}
}

func RandomNetworkingConfig() ConfigureNode {
	return func() config.Config {
		cfg := config.DefaultConfig()
		cfg.Net.P2PAddress = "/ip4/0.0.0.0/tcp/0"
		cfg.Net.RelayEnabled = false
		return *cfg
	}
}
