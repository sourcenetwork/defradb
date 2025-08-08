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

	netConfig "github.com/sourcenetwork/defradb/net/config"
	"github.com/sourcenetwork/defradb/tests/state"

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

	// NonExistentDocID can be used to represent a non-existent docID, it will be substituted
	// for a non-existent dicID when used in actions that support this.
	NonExistentDocID       int    = -1
	NonExistentDocIDString string = "NonExistentDocID"
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

// SubscribeToDocument sets up a subscription on the given node to the given document.
//
// Changes made to subscribed documents in peers connected to this node will be synced from
// them to this node.
type SubscribeToDocument struct {
	// NodeID is the node ID (index) of the node in which to activate the subscription.
	//
	// Changes made to subscribed documents in peers connected to this node will be synced from
	// them to this node.
	NodeID int

	// DocIDs are the docIDs (indexes) of the documents to subscribe to.
	//
	// A [NonExistentDocID] may be provided to test non-existent  docIDs.
	DocIDs []state.ColDocIndex

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// UnsubscribeToDocument removes the given documents from the set of active subscriptions on
// the given node.
type UnsubscribeToDocument struct {
	// NodeID is the node ID (index) of the node in which to remove the subscription.
	NodeID int

	// DocIDs are the docIDs (indexes) of the documents to unsubscribe from.
	//
	// A [NonExistentDocID] may be provided to test non-existent docIDs.
	DocIDs []state.ColDocIndex

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// GetAllP2PDocuments gets the active subscriptions for the given node and compares them against the
// expected results.
type GetAllP2PDocuments struct {
	// NodeID is the node ID (index) of the node in which to get the subscriptions for.
	NodeID int

	// ExpectedDocIDs are the docIDs (indexes) of the documents expected.
	ExpectedDocIDs []state.ColDocIndex
}

// WaitForSync is an action that instructs the test framework to wait for all document synchronization
// to complete before progressing.
//
// For example you will likely wish to `WaitForSync` after creating a document in node 0 before querying
// node 1 to see if it has been replicated.
type WaitForSync struct {
	// Decrypted is a list of document indexes that are expected to be merged and synced decrypted.
	Decrypted []int
}

// connectPeers connects two existing, started, nodes as peers.  It returns a channel
// that will receive an empty struct upon sync completion of all expected peer-sync events.
//
// Any errors generated whilst configuring the peers or waiting on sync will result in a test failure.
func connectPeers(
	s *state.State,
	cfg ConnectPeers,
) {
	sourceNode := s.Nodes[cfg.SourceNodeID]
	targetNode := s.Nodes[cfg.TargetNodeID]

	log.InfoContext(s.Ctx, "Connect peers",
		corelog.Any("Source", sourceNode.PeerInfo()),
		corelog.Any("Target", targetNode.PeerInfo()))

	err := sourceNode.Connect(s.Ctx, targetNode.PeerInfo())
	require.NoError(s.T, err)

	s.Nodes[cfg.SourceNodeID].P2P.Connections[cfg.TargetNodeID] = struct{}{}
	s.Nodes[cfg.TargetNodeID].P2P.Connections[cfg.SourceNodeID] = struct{}{}

	// Bootstrap triggers a bunch of async stuff for which we have no good way of waiting on.  It must be
	// allowed to complete before documentation begins or it will not even try and sync it. So for now, we
	// sleep a little.
	time.Sleep(10 * time.Millisecond)
}

// configureReplicator configures a replicator relationship between two existing, started, nodes.
// It returns a channel that will receive an empty struct upon sync completion of all expected
// replicator-sync events.
//
// Any errors generated whilst configuring the peers or waiting on sync will result in a test failure.
func configureReplicator(
	s *state.State,
	cfg ConfigureReplicator,
) {
	sourceNode := s.Nodes[cfg.SourceNodeID]
	targetNode := s.Nodes[cfg.TargetNodeID]

	err := sourceNode.SetReplicator(s.Ctx, targetNode.PeerInfo())

	expectedErrorRaised := AssertError(s.T, err, cfg.ExpectedError)
	assertExpectedErrorRaised(s.T, cfg.ExpectedError, expectedErrorRaised)

	if err == nil {
		waitForReplicatorConfigureEvent(s, cfg)
	}
}

func deleteReplicator(
	s *state.State,
	cfg DeleteReplicator,
) {
	sourceNode := s.Nodes[cfg.SourceNodeID]
	targetNode := s.Nodes[cfg.TargetNodeID]

	err := sourceNode.DeleteReplicator(s.Ctx, targetNode.PeerInfo())
	require.NoError(s.T, err)
	waitForReplicatorDeleteEvent(s, cfg)
}

// subscribeToCollection sets up a collection subscription on the given node/collection.
//
// Any errors generated during this process will result in a test failure.
func subscribeToCollection(
	s *state.State,
	action SubscribeToCollection,
) {
	n := s.Nodes[action.NodeID]

	collectionNames := []string{}
	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			collectionNames = append(collectionNames, NonExistentCollectionSchemaRoot)
			continue
		}

		col := s.Nodes[action.NodeID].Collections[collectionIndex]
		collectionNames = append(collectionNames, col.Name())
	}

	err := n.AddP2PCollections(s.Ctx, collectionNames...)
	if err == nil {
		waitForSubscribeToCollectionEvent(s, action)
	}

	expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)
	assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

	// The `n.Peer.AddP2PCollections(colIDs)` call above is calling some asynchronous functions
	// for the pubsub subscription and those functions can take a bit of time to complete,
	// we need to make sure this has finished before progressing.
	time.Sleep(100 * time.Millisecond)
}

// unsubscribeToCollection removes the given collections from subscriptions on the given nodes.
//
// Any errors generated during this process will result in a test failure.
func unsubscribeToCollection(
	s *state.State,
	action UnsubscribeToCollection,
) {
	n := s.Nodes[action.NodeID]

	collectionNames := []string{}
	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			collectionNames = append(collectionNames, NonExistentCollectionSchemaRoot)
			continue
		}

		col := s.Nodes[action.NodeID].Collections[collectionIndex]
		collectionNames = append(collectionNames, col.Name())
	}

	err := n.RemoveP2PCollections(s.Ctx, collectionNames...)
	if err == nil {
		waitForUnsubscribeToCollectionEvent(s, action)
	}

	expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)
	assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

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
	s *state.State,
	action GetAllP2PCollections,
) {
	expectedCollections := []string{}
	for _, collectionIndex := range action.ExpectedCollectionIDs {
		col := s.Nodes[action.NodeID].Collections[collectionIndex]
		expectedCollections = append(expectedCollections, col.Name())
	}

	n := s.Nodes[action.NodeID]
	cols, err := n.GetAllP2PCollections(s.Ctx)
	require.NoError(s.T, err)

	assert.Equal(s.T, expectedCollections, cols)
}

// subscribeToDocument sets up a collection subscription on the given node/collection.
//
// Any errors generated during this process will result in a test failure.
func subscribeToDocument(
	s *state.State,
	action SubscribeToDocument,
) {
	n := s.Nodes[action.NodeID]

	docIDs := []string{}
	for _, colDocIndex := range action.DocIDs {
		if colDocIndex.Doc == NonExistentDocID {
			docIDs = append(docIDs, NonExistentDocIDString)
			continue
		}

		docID := s.DocIDs[colDocIndex.Col][colDocIndex.Doc]
		docIDs = append(docIDs, docID.String())
	}

	err := n.AddP2PDocuments(s.Ctx, docIDs...)
	if err == nil {
		waitForSubscribeToDocumentEvent(s, action)
	}

	expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)
	assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

	// The `n.Peer.AddP2PDocuments(colIDs)` call above is calling some asynchronous functions
	// for the pubsub subscription and those functions can take a bit of time to complete,
	// we need to make sure this has finished before progressing.
	time.Sleep(100 * time.Millisecond)
}

// unsubscribeToDocument removes the given collections from subscriptions on the given nodes.
//
// Any errors generated during this process will result in a test failure.
func unsubscribeToDocument(
	s *state.State,
	action UnsubscribeToDocument,
) {
	n := s.Nodes[action.NodeID]

	docIDs := []string{}
	for _, colDocIndex := range action.DocIDs {
		if colDocIndex.Doc == NonExistentDocID {
			docIDs = append(docIDs, NonExistentDocIDString)
			continue
		}

		docID := s.DocIDs[colDocIndex.Col][colDocIndex.Doc]
		docIDs = append(docIDs, docID.String())
	}

	err := n.RemoveP2PDocuments(s.Ctx, docIDs...)
	if err == nil {
		waitForUnsubscribeToDocumentEvent(s, action)
	}

	expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)
	assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

	// The `n.Peer.RemoveP2PDocuments(colIDs)` call above is calling some asynchronous functions
	// for the pubsub subscription and those functions can take a bit of time to complete,
	// we need to make sure this has finished before progressing.
	time.Sleep(100 * time.Millisecond)
}

// getAllP2PDocuments gets all the active peer subscriptions and compares them against the
// given expected results.
//
// Any errors generated during this process will result in a test failure.
func getAllP2PDocuments(
	s *state.State,
	action GetAllP2PDocuments,
) {
	expectedDocuments := []string{}
	for _, colDocIndex := range action.ExpectedDocIDs {
		docID := s.DocIDs[colDocIndex.Col][colDocIndex.Doc]
		expectedDocuments = append(expectedDocuments, docID.String())
	}

	n := s.Nodes[action.NodeID]
	cols, err := n.GetAllP2PDocuments(s.Ctx)
	require.NoError(s.T, err)

	assert.Equal(s.T, expectedDocuments, cols)
}

// reconnectPeers makes sure that all peers are connected after a node restart action.
func reconnectPeers(s *state.State) {
	for i, n := range s.Nodes {
		for j := range n.P2P.Connections {
			sourceNode := s.Nodes[i]
			targetNode := s.Nodes[j]

			log.InfoContext(s.Ctx, "Connect peers",
				corelog.Any("Source", sourceNode.PeerInfo()),
				corelog.Any("Target", targetNode.PeerInfo()))

			err := sourceNode.Connect(s.Ctx, targetNode.PeerInfo())
			require.NoError(s.T, err)
		}
	}
}

func RandomNetworkingConfig() ConfigureNode {
	return func() []netConfig.NodeOpt {
		return []netConfig.NodeOpt{
			netConfig.WithListenAddresses("/ip4/127.0.0.1/tcp/0"),
			netConfig.WithEnableRelay(false),
		}
	}
}

// syncDocs requests document sync from peers.
func syncDocs(s *state.State, action SyncDocs) {
	node := s.Nodes[action.NodeID]

	docIDStrings := make([]string, len(action.DocIDs))
	for i, docIndex := range action.DocIDs {
		docIDStrings[i] = s.DocIDs[action.CollectionID][docIndex].String()
	}

	collectionName := s.Nodes[action.NodeID].Collections[action.CollectionID].Name()

	err := withRetryOnNode(
		node,
		func() error {
			return node.SyncDocuments(
				s.Ctx,
				collectionName,
				docIDStrings,
			)
		},
	)

	expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)

	assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

	if !expectedErrorRaised {
		for i, docInd := range action.DocIDs {
			nodeID := action.SourceNodes[i]
			docID := s.DocIDs[action.CollectionID][docInd].String()
			node.P2P.ExpectedDAGHeads[docID] = s.Nodes[nodeID].P2P.ActualDAGHeads[docID].CID
		}
	}
}
