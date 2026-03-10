// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package tests

import (
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

const (
	// NonExistentCollectionID can be used to represent a non-existent collection ID, it will be substituted
	// for a non-existent collection ID when used in actions that support this.
	NonExistentCollectionID   int    = -1
	NonExistentCollectionRoot string = "NonExistentCollectionRoot"
)

// AddCollectionSubscription sets up a subscription on the given node to the given collection.
//
// Changes made to subscribed collections in peers connected to this node will be synced from
// them to this node.
type AddCollectionSubscription struct {
	// NodeID is the node ID (index) of the node in which to activate the subscription.
	//
	// Changes made to subscribed collections in peers connected to this node will be synced from
	// them to this node.
	NodeID int

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

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

// DeleteCollectionSubscription removes the given collections from the set of active subscriptions on
// the given node.
type DeleteCollectionSubscription struct {
	// NodeID is the node ID (index) of the node in which to remove the subscription.
	NodeID int

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

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

// ListP2PCollections gets the active subscriptions for the given node and compares them against the
// expected results.
type ListP2PCollections struct {
	// NodeID is the node ID (index) of the node in which to get the subscriptions for.
	NodeID int

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// ExpectedCollectionIDs are the collection IDs (indexes) of the collections expected.
	ExpectedCollectionIDs []int

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// addCollectionSubscription sets up a collection subscription on the given node/collection.
//
// Any errors generated during this process will result in a test failure.
func addCollectionSubscription(
	s *state.State,
	action AddCollectionSubscription,
) {
	node := s.Nodes[action.NodeID]

	collectionNames := []string{}
	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			collectionNames = append(collectionNames, NonExistentCollectionRoot)
			continue
		}

		col := s.Nodes[action.NodeID].Collections[collectionIndex]
		collectionNames = append(collectionNames, col.Name())
	}

	opt := options.WithIdentity(options.AddP2PCollections(),
		getIdentityForRequestSpecificToNode(s, action.Identity, action.NodeID))
	err := node.AddP2PCollections(s.Ctx, collectionNames, opt)
	if err == nil {
		waitForAddCollectionSubscriptionEvent(s, action)
	}

	expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)
	assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

	// The `n.Peer.AddP2PCollections(colIDs)` call above is calling some asynchronous functions
	// for the pubsub subscription and those functions can take a bit of time to complete,
	// we need to make sure this has finished before progressing.
	time.Sleep(100 * time.Millisecond)
}

// deleteCollectionSubscription removes the given collections from subscriptions on the given nodes.
//
// Any errors generated during this process will result in a test failure.
func deleteCollectionSubscription(
	s *state.State,
	action DeleteCollectionSubscription,
) {
	node := s.Nodes[action.NodeID]

	collectionNames := []string{}
	for _, collectionIndex := range action.CollectionIDs {
		if collectionIndex == NonExistentCollectionID {
			collectionNames = append(collectionNames, NonExistentCollectionRoot)
			continue
		}

		col := s.Nodes[action.NodeID].Collections[collectionIndex]
		collectionNames = append(collectionNames, col.Name())
	}

	opt := options.WithIdentity(options.DeleteP2PCollections(),
		getIdentityForRequestSpecificToNode(s, action.Identity, action.NodeID))
	err := node.DeleteP2PCollections(s.Ctx, collectionNames, opt)
	if err == nil {
		waitForDeleteCollectionSubscriptionEvent(s, action)
	}

	expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)
	assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

	// The `n.Peer.RemoveP2PCollections(colIDs)` call above is calling some asynchronous functions
	// for the pubsub subscription and those functions can take a bit of time to complete,
	// we need to make sure this has finished before progressing.
	time.Sleep(100 * time.Millisecond)
}

// listP2PCollections gets all the active peer subscriptions and compares them against the
// given expected results.
//
// Any errors generated during this process will result in a test failure.
func listP2PCollections(
	s *state.State,
	action ListP2PCollections,
) {
	expectedCollections := []string{}
	for _, collectionIndex := range action.ExpectedCollectionIDs {
		col := s.Nodes[action.NodeID].Collections[collectionIndex]
		expectedCollections = append(expectedCollections, col.Name())
	}

	node := s.Nodes[action.NodeID]
	opt := options.WithIdentity(options.ListP2PCollections(),
		getIdentityForRequestSpecificToNode(s, action.Identity, action.NodeID))
	cols, err := node.ListP2PCollections(s.Ctx, opt)

	expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)
	assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

	if !expectedErrorRaised {
		assert.Equal(s.T, expectedCollections, cols)
	}
}
