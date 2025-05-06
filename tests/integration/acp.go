// Copyright 2024 Democratized Data Foundation
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
	"os"
	"slices"
	"time"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"
)

type DocumentACPType string

const (
	documentACPTypeEnvName = "DEFRA_DOCUMENT_ACP_TYPE"
)

const (
	SourceHubDocumentACPType DocumentACPType = "source-hub"
	LocalDocumentACPType     DocumentACPType = "local"
)

const (
	// authTokenExpiration is the expiration time for auth tokens.
	authTokenExpiration = time.Minute * 1
)

var (
	documentACPType DocumentACPType
)

// KMSType is the type of KMS to use.
type KMSType string

const (
	// NoneKMSType is the none KMS type. It is used to indicate that no KMS should be used.
	NoneKMSType KMSType = "none"
	// PubSubKMSType is the PubSub KMS type.
	PubSubKMSType KMSType = "pubsub"
)

func getKMSTypes() []KMSType {
	return []KMSType{PubSubKMSType}
}

func init() {
	documentACPType = DocumentACPType(os.Getenv(documentACPTypeEnvName))
	if documentACPType == "" {
		documentACPType = LocalDocumentACPType
	}
}

// AddDocPolicy will attempt to add the given policy using DefraDB's Document ACP system.
type AddDocPolicy struct {
	// NodeID may hold the ID (index) of the node we want to add policy to.
	//
	// If a value is not provided the policy will be added in all nodes, unless testing with
	// sourcehub ACP, in which case the policy will only be defined once.
	NodeID immutable.Option[int]

	// The raw policy string.
	Policy string

	// The policy creator identity, i.e. actor creating the policy.
	Identity immutable.Option[Identity]

	// The expected policyID generated based on the Policy loaded in to the ACP system.
	//
	// This is an optional attribute, for situations where a test might want to assert
	// the exact policyID. When this is not provided the test will just assert that
	// the resulting policyID is not empty.
	ExpectedPolicyID immutable.Option[string]

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// addPolicyDocumentACP will attempt to add the given policy using DefraDB's Document ACP system.
func addPolicyDocumentACP(
	s *state,
	action AddDocPolicy,
) {
	// If we expect an error, then ExpectedPolicyID should never be provided.
	if action.ExpectedError != "" && action.ExpectedPolicyID.HasValue() {
		require.Fail(s.t, "Expected error should not have an expected policyID with it.", s.testCase.Description)
	}

	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	maxNodeID := slices.Max(nodeIDs)
	// Expand the policyIDs slice once, so we can minimize how many times we need to expaind it.
	// We use the maximum nodeID provided to make sure policyIDs slice can accomodate upto that nodeID.
	if len(s.policyIDs) <= maxNodeID {
		// Expand the slice if required, so that the policyID can be accessed by node index
		policyIDs := make([][]string, maxNodeID+1)
		copy(policyIDs, s.policyIDs)
		s.policyIDs = policyIDs
	}

	for index, node := range nodes {
		nodeID := nodeIDs[index]
		ctx := getContextWithIdentity(s.ctx, s, action.Identity, nodeID)
		policyResult, err := node.AddPolicy(ctx, action.Policy)

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(s.t, action.ExpectedError, "")
			if action.ExpectedPolicyID.HasValue() {
				require.Equal(s.t, action.ExpectedPolicyID.Value(), policyResult.PolicyID)
			} else {
				require.NotEqual(s.t, policyResult.PolicyID, "")
			}

			s.policyIDs[nodeID] = append(s.policyIDs[nodeID], policyResult.PolicyID)
		}

		// The policy should only be added to a SourceHub chain once - there is no need to loop through
		// the nodes.
		if documentACPType == SourceHubDocumentACPType {
			// Note: If we break here the state will only preserve the policyIDs result on the
			// first node if acp type is sourcehub, make sure to replicate the policyIDs state
			// on all the nodes, so we don't have to handle all the edge cases later in actions.
			for otherIndexes := index + 1; otherIndexes < len(nodes); otherIndexes++ {
				s.policyIDs[nodeIDs[otherIndexes]] = s.policyIDs[nodeID]
			}
			break
		}
	}
}

// AddDocActorRelationship will attempt to create a new relationship for a document with an actor.
type AddDocActorRelationship struct {
	// NodeID may hold the ID (index) of the node we want to add doc actor relationship on.
	//
	// If a value is not provided the relationship will be added in all nodes, unless testing with
	// sourcehub ACP, in which case the relationship will only be defined once.
	NodeID immutable.Option[int]

	// The collection in which this document we want to add a relationship for exists.
	//
	// This is a required field. To test the invalid usage of not having this arg, use -1 index.
	CollectionID int

	// The index-identifier of the document within the collection.  This is based on
	// the order in which it was created, not the ordering of the document within the
	// database.
	//
	// This is a required field. To test the invalid usage of not having this arg, use -1 index.
	DocID int

	// The name of the relation to set between document and target actor (should be defined in the policy).
	//
	// This is a required field.
	Relation string

	// The target public identity, i.e. the identity of the actor to tie the document's relation with.
	//
	// This is a required field. To test the invalid usage of not having this arg, use NoIdentity() or leave default.
	TargetIdentity immutable.Option[Identity]

	// The requestor identity, i.e. identity of the actor creating the relationship.
	// Note: This identity must either own or have managing access defined in the policy.
	//
	// This is a required field. To test the invalid usage of not having this arg, use NoIdentity() or leave default.
	RequestorIdentity immutable.Option[Identity]

	// Result returns true if it was a no-op due to existing before, and false if a new relationship was made.
	ExpectedExistence bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

func addDocActorRelationshipACP(
	s *state,
	action AddDocActorRelationship,
) {
	var docID string
	actionNodeID := action.NodeID
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		var collectionName string
		collectionName, docID = getCollectionAndDocInfo(s, action.CollectionID, action.DocID, nodeID)

		exists, err := node.AddDocActorRelationship(
			getContextWithIdentity(s.ctx, s, action.RequestorIdentity, nodeID),
			collectionName,
			docID,
			action.Relation,
			getIdentityDID(s, action.TargetIdentity),
		)

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(s.t, action.ExpectedError, "")
			require.Equal(s.t, action.ExpectedExistence, exists.ExistedAlready)
		}

		// The relationship should only be added to a SourceHub chain once - there is no need to loop through
		// the nodes.
		if documentACPType == SourceHubDocumentACPType {
			actionNodeID = immutable.Some(0)
			break
		}
	}

	if action.ExpectedError == "" && !action.ExpectedExistence {
		expect := map[string]struct{}{
			docID: {},
		}

		waitForUpdateEvents(s, actionNodeID, action.CollectionID, expect, action.TargetIdentity)
	}
}

// DeleteDocActorRelationship will attempt to delete a relationship between a document and an actor.
type DeleteDocActorRelationship struct {
	// NodeID may hold the ID (index) of the node we want to delete doc actor relationship on.
	//
	// If a value is not provided the relationship will be deleted on all nodes, unless testing with
	// sourcehub document ACP, in which case the relationship will only be deleted once.
	NodeID immutable.Option[int]

	// The collection in which the target document we want to delete relationship for exists.
	//
	// This is a required field. To test the invalid usage of not having this arg, use -1 index.
	CollectionID int

	// The index-identifier of the document within the collection.  This is based on
	// the order in which it was created, not the ordering of the document within the
	// database.
	//
	// This is a required field. To test the invalid usage of not having this arg, use -1 index.
	DocID int

	// The name of the relation within the relationship we want to delete (should be defined in the policy).
	//
	// This is a required field.
	Relation string

	// The target public identity, i.e. the identity of the actor with whom the relationship is with.
	//
	// This is a required field. To test the invalid usage of not having this arg, use NoIdentity() or leave default.
	TargetIdentity immutable.Option[Identity]

	// The requestor identity, i.e. identity of the actor deleting the relationship.
	// Note: This identity must either own or have managing access defined in the policy.
	//
	// This is a required field. To test the invalid usage of not having this arg, use NoIdentity() or leave default.
	RequestorIdentity immutable.Option[Identity]

	// Result returns true if the relationship record was expected to be found and deleted,
	// and returns false if no matching relationship record was found (no-op).
	ExpectedRecordFound bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

func deleteDocActorRelationshipACP(
	s *state,
	action DeleteDocActorRelationship,
) {
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		collectionName, docID := getCollectionAndDocInfo(s, action.CollectionID, action.DocID, nodeID)

		deleteDocActorRelationshipResult, err := node.DeleteDocActorRelationship(
			getContextWithIdentity(s.ctx, s, action.RequestorIdentity, nodeID),
			collectionName,
			docID,
			action.Relation,
			getIdentityDID(s, action.TargetIdentity),
		)

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(s.t, action.ExpectedError, "")
			require.Equal(s.t, action.ExpectedRecordFound, deleteDocActorRelationshipResult.RecordFound)
		}

		// The relationship should only be added to a SourceHub chain once - there is no need to loop through
		// the nodes.
		if documentACPType == SourceHubDocumentACPType {
			break
		}
	}
}

func getCollectionAndDocInfo(s *state, collectionID, docInd, nodeID int) (string, string) {
	collectionName := ""
	docID := ""
	if collectionID != -1 {
		collection := s.nodes[nodeID].collections[collectionID]
		if collection.Description().Name == "" {
			require.Fail(s.t, "Expected non-empty collection name, but it was empty.", s.testCase.Description)
		}
		collectionName = collection.Description().Name

		if docInd != -1 {
			docID = s.docIDs[collectionID][docInd].String()
		}
	}
	return collectionName, docID
}
