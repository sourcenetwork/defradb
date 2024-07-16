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
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/net"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sourcenetwork/corelog"
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

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
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
	NonExistentCollectionID         int    = -1
	NonExistentCollectionSchemaRoot string = "NonExistentCollectionID"
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
	log.InfoContext(s.ctx, "Bootstrapping with peers", corelog.Any("Addresses", addrs))
	sourceNode.Bootstrap(addrs)

	s.nodeConnections[cfg.SourceNodeID][cfg.TargetNodeID] = struct{}{}
	s.nodeConnections[cfg.TargetNodeID][cfg.SourceNodeID] = struct{}{}

	// Bootstrap triggers a bunch of async stuff for which we have no good way of waiting on.  It must be
	// allowed to complete before documentation begins or it will not even try and sync it. So for now, we
	// sleep a little.
	time.Sleep(100 * time.Millisecond)
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

	sub, err := sourceNode.Events().Subscribe(event.ReplicatorCompletedName)
	require.NoError(s.t, err)
	err = sourceNode.SetReplicator(s.ctx, client.Replicator{
		Info: targetNode.PeerInfo(),
	})
	if err == nil {
		// wait for the replicator setup to complete
		<-sub.Message()
	}

	expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, cfg.ExpectedError)
	assertExpectedErrorRaised(s.t, s.testCase.Description, cfg.ExpectedError, expectedErrorRaised)

	if err == nil {
		// all previous documents should be merged on the subscriber node
		for key, val := range s.actualDocHeads[cfg.SourceNodeID] {
			s.expectedDocHeads[cfg.TargetNodeID][key] = val
		}
		s.nodeReplicatorTargets[cfg.TargetNodeID][cfg.SourceNodeID] = struct{}{}
		s.nodeReplicatorSources[cfg.SourceNodeID][cfg.TargetNodeID] = struct{}{}
	}
}

func deleteReplicator(
	s *state,
	cfg DeleteReplicator,
) {
	sourceNode := s.nodes[cfg.SourceNodeID]
	targetNode := s.nodes[cfg.TargetNodeID]

	sub, err := sourceNode.Events().Subscribe(event.ReplicatorCompletedName)
	require.NoError(s.t, err)
	err = sourceNode.DeleteReplicator(s.ctx, client.Replicator{
		Info: targetNode.PeerInfo(),
	})
	if err == nil {
		// wait for the replicator setup to complete
		<-sub.Message()
	}
	require.NoError(s.t, err)
	delete(s.nodeReplicatorTargets[cfg.TargetNodeID], cfg.SourceNodeID)
	delete(s.nodeReplicatorSources[cfg.SourceNodeID], cfg.TargetNodeID)
}

// subscribeToCollection sets up a collection subscription on the given node/collection.
//
// Any errors generated during this process will result in a test failure.
func subscribeToCollection(
	s *state,
	action SubscribeToCollection,
) {
	n := s.nodes[action.NodeID]

	schemaRoots := []string{}
	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			schemaRoots = append(schemaRoots, NonExistentCollectionSchemaRoot)
			continue
		}
		if action.ExpectedError == "" {
			// all previous documents should be merged on the subscriber node
			if collectionIndex < len(s.documents) {
				for _, doc := range s.documents[collectionIndex] {
					for nodeID := range s.nodeConnections[action.NodeID] {
						s.expectedDocHeads[action.NodeID][doc.ID().String()] = s.actualDocHeads[nodeID][doc.ID().String()]
					}
				}
			}
			s.nodePeerCollections[collectionIndex][action.NodeID] = struct{}{}
		}
		col := s.collections[action.NodeID][collectionIndex]
		schemaRoots = append(schemaRoots, col.SchemaRoot())
	}

	sub, err := n.Events().Subscribe(event.P2PTopicCompletedName)
	require.NoError(s.t, err)

	err = n.AddP2PCollections(s.ctx, schemaRoots)
	if err == nil {
		// wait for the p2p collection setup to complete
		<-sub.Message()
	}

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

	schemaRoots := []string{}
	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			schemaRoots = append(schemaRoots, NonExistentCollectionSchemaRoot)
			continue
		}
		if action.ExpectedError == "" {
			delete(s.nodePeerCollections[collectionIndex], action.NodeID)
		}
		col := s.collections[action.NodeID][collectionIndex]
		schemaRoots = append(schemaRoots, col.SchemaRoot())
	}

	sub, err := n.Events().Subscribe(event.P2PTopicCompletedName)
	require.NoError(s.t, err)

	err = n.RemoveP2PCollections(s.ctx, schemaRoots)
	if err == nil {
		// wait for the p2p collection setup to complete
		<-sub.Message()
	}

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
		expectedCollections = append(expectedCollections, col.SchemaRoot())
	}

	n := s.nodes[action.NodeID]
	cols, err := n.GetAllP2PCollections(s.ctx)
	require.NoError(s.t, err)

	assert.Equal(s.t, expectedCollections, cols)
}

func RandomNetworkingConfig() ConfigureNode {
	return func() []net.NodeOpt {
		return []net.NodeOpt{
			net.WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
			net.WithEnableRelay(false),
		}
	}
}
