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
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

// ReEnableAAC will attempt to re-enable a temporarily disabled Admin ACP system.
type ReEnableAAC struct {
	// NodeID may hold the ID (index) of the node we want to re-enable the admin acp on.
	//
	// If a value is not provided, then will start admin acp on all nodes.
	NodeID immutable.Option[int]

	// The identity of the user that is re-enabling admin acp, this user must be authorized
	// to re-enable admin acp, otherwise error will be returned.
	Identity immutable.Option[Identity]

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// reEnableAAC will attempt to re-enable the admin access control system.
func reEnableAAC(
	s *state,
	action ReEnableAAC,
) {
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		ctx := getContextWithIdentity(s.ctx, s, action.Identity, nodeID)
		err := node.ReEnableAAC(ctx)

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
		if !expectedErrorRaised {
			require.Equal(s.t, action.ExpectedError, "")
		}
	}
}

// DisableAAC will attempt to temporarily disable DefraDB's Admin ACP system.
type DisableAAC struct {
	// NodeID may hold the ID (index) of the node we want to disable the admin acp on.
	//
	// If a value is not provided, then will disable admin acp on all nodes.
	NodeID immutable.Option[int]

	// The identity of a user that is authorized to disable the admin acp.
	// The identity of the user that is disabling admin acp, this user must be authorized
	// to disable admin acp, otherwise error will be returned.
	Identity immutable.Option[Identity]

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// disableAAC will attempt to start the admin access control system.
func disableAAC(
	s *state,
	action DisableAAC,
) {
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		ctx := getContextWithIdentity(s.ctx, s, action.Identity, nodeID)
		err := node.DisableAAC(ctx)

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
		if !expectedErrorRaised {
			require.Equal(s.t, action.ExpectedError, "")
		}
	}
}

// AddAACActorRelationship will attempt to create a new relationship for node access with an actor.
type AddAACActorRelationship struct {
	// NodeID may hold the ID (index) of the node we want to add actor relationship on.
	//
	// If a value is not provided the relationship will be added in all nodes.
	NodeID immutable.Option[int]

	// The name of the relation to set for target actor (MUST be defined in the uploaded admin policy).
	//
	// This is a required field.
	Relation string

	// The target public identity, i.e. the identity of the actor to create a relationship with.
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

func addAACActorRelationship(
	s *state,
	action AddAACActorRelationship,
) {
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		addActorRelationshipResult, err := node.AddAACActorRelationship(
			getContextWithIdentity(s.ctx, s, action.RequestorIdentity, nodeID),
			action.Relation,
			getIdentityDID(s, action.TargetIdentity),
		)

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(s.t, action.ExpectedError, "")
			require.Equal(s.t, action.ExpectedExistence, addActorRelationshipResult.ExistedAlready)
		}
	}
}

// DeleteAACActorRelationship will attempt to delete a relationship for node access with an actor.
type DeleteAACActorRelationship struct {
	// NodeID may hold the ID (index) of the node we want to delete actor relationship on.
	//
	// If a value is not provided the relationship will be deleted on all nodes.
	NodeID immutable.Option[int]

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

func deleteAACActorRelationship(
	s *state,
	action DeleteAACActorRelationship,
) {
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		deleteActorRelationshipResult, err := node.DeleteAACActorRelationship(
			getContextWithIdentity(s.ctx, s, action.RequestorIdentity, nodeID),
			action.Relation,
			getIdentityDID(s, action.TargetIdentity),
		)

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(s.t, action.ExpectedError, "")
			require.Equal(s.t, action.ExpectedRecordFound, deleteActorRelationshipResult.RecordFound)
		}
	}
}

// GetAACStatus returns the current status of the target node(s) if all goes well,
// otherwise returns an error.
type GetAACStatus struct {
	// NodeID may hold the ID (index) of the node we want to check admin acp status on.
	//
	// If a value is not provided, then will check admin acp status on all nodes.
	NodeID immutable.Option[int]

	// The identity of the user that is making this request.
	Identity immutable.Option[Identity]

	// ExpectedStatus returns the current status of the NAC system.
	ExpectedStatus client.NACStatus

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

func getAACStatus(
	s *state,
	action GetAACStatus,
) {
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		statusAACResult, err := node.GetAACStatus(
			getContextWithIdentity(s.ctx, s, action.Identity, nodeID),
		)

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(s.t, action.ExpectedError, "")
			require.Equal(s.t, action.ExpectedStatus.String(), statusAACResult.Status)
		}
	}
}
