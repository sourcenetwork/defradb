// Copyright 2025 Democratized Data Foundation
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

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/state"
)

const (
	// NonExistentDocID can be used to represent a non-existent docID, it will be substituted
	// for a non-existent dicID when used in actions that support this.
	NonExistentDocID       int    = -1
	NonExistentDocIDString string = "NonExistentDocID"
)

// CreateDocumentSubscription sets up a subscription on the given node to the given document.
//
// Changes made to subscribed documents in peers connected to this node will be synced from
// them to this node.
type CreateDocumentSubscription struct {
	// NodeID is the node ID (index) of the node in which to activate the subscription.
	//
	// Changes made to subscribed documents in peers connected to this node will be synced from
	// them to this node.
	NodeID int

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

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

// DeleteDocumentSubscription removes the given documents from the set of active subscriptions on
// the given node.
type DeleteDocumentSubscription struct {
	// NodeID is the node ID (index) of the node in which to remove the subscription.
	NodeID int

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

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

// ListP2PDocuments gets the active subscriptions for the given node and compares them against the
// expected results.
type ListP2PDocuments struct {
	// NodeID is the node ID (index) of the node in which to get the subscriptions for.
	NodeID int

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// ExpectedDocIDs are the docIDs (indexes) of the documents expected.
	ExpectedDocIDs []state.ColDocIndex

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// createDocumentSubscription sets up a collection subscription on the given node/collection.
//
// Any errors generated during this process will result in a test failure.
func createDocumentSubscription(
	s *state.State,
	action CreateDocumentSubscription,
) {
	node := s.Nodes[action.NodeID]

	docIDs := []string{}
	for _, colDocIndex := range action.DocIDs {
		if colDocIndex.Doc == NonExistentDocID {
			docIDs = append(docIDs, NonExistentDocIDString)
			continue
		}

		s.DocIDsLock.RLock()
		docID := s.DocIDs[colDocIndex.Col][colDocIndex.Doc]
		s.DocIDsLock.RUnlock()

		docIDs = append(docIDs, docID.String())
	}

	ctx := getContextWithIdentity(s.Ctx, s, action.Identity, action.NodeID)
	err := node.CreateP2PDocuments(ctx, docIDs...)
	if err == nil {
		waitForCreateDocumentSubscriptionEvent(s, action)
	}

	expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)
	assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

	// The `n.Peer.CreateP2PDocuments(colIDs)` call above is calling some asynchronous functions
	// for the pubsub subscription and those functions can take a bit of time to complete,
	// we need to make sure this has finished before progressing.
	time.Sleep(100 * time.Millisecond)
}

// deleteDocumentSubscription removes the given collections from subscriptions on the given nodes.
//
// Any errors generated during this process will result in a test failure.
func deleteDocumentSubscription(
	s *state.State,
	action DeleteDocumentSubscription,
) {
	node := s.Nodes[action.NodeID]

	docIDs := []string{}
	for _, colDocIndex := range action.DocIDs {
		if colDocIndex.Doc == NonExistentDocID {
			docIDs = append(docIDs, NonExistentDocIDString)
			continue
		}

		s.DocIDsLock.RLock()
		docID := s.DocIDs[colDocIndex.Col][colDocIndex.Doc]
		s.DocIDsLock.RUnlock()

		docIDs = append(docIDs, docID.String())
	}

	ctx := getContextWithIdentity(s.Ctx, s, action.Identity, action.NodeID)
	err := node.DeleteP2PDocuments(ctx, docIDs...)
	if err == nil {
		waitForDeleteDocumentSubscriptionEvent(s, action)
	}

	expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)
	assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

	// The `n.Peer.DeleteP2PDocuments(colIDs)` call above is calling some asynchronous functions
	// for the pubsub subscription and those functions can take a bit of time to complete,
	// we need to make sure this has finished before progressing.
	time.Sleep(100 * time.Millisecond)
}

// listP2PDocuments gets all the active peer subscriptions and compares them against the
// given expected results.
//
// Any errors generated during this process will result in a test failure.
func listP2PDocuments(
	s *state.State,
	action ListP2PDocuments,
) {
	expectedDocuments := []string{}
	for _, colDocIndex := range action.ExpectedDocIDs {
		s.DocIDsLock.RLock()
		docID := s.DocIDs[colDocIndex.Col][colDocIndex.Doc]
		s.DocIDsLock.RUnlock()

		expectedDocuments = append(expectedDocuments, docID.String())
	}

	node := s.Nodes[action.NodeID]
	ctx := getContextWithIdentity(s.Ctx, s, action.Identity, action.NodeID)
	cols, err := node.ListP2PDocuments(ctx)

	expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)
	assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

	if !expectedErrorRaised {
		assert.Equal(s.T, expectedDocuments, cols)
	}
}
