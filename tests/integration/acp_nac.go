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
	"github.com/sourcenetwork/defradb/tests/state"
)

// ReEnableNAC will attempt to re-enable a temporarily disabled Node ACP system.
type ReEnableNAC struct {
	// NodeID may hold the ID (index) of the node we want to re-enable the node acp on.
	//
	// If a value is not provided, then will start node acp on all nodes.
	NodeID immutable.Option[int]

	// The identity of the user that is re-enabling node acp, this user must be authorized
	// to re-enable node acp, otherwise error will be returned.
	Identity immutable.Option[state.Identity]

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// reEnableNAC will attempt to re-enable the node access control system.
func reEnableNAC(
	s *state.State,
	action ReEnableNAC,
) {
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		ctx := getContextWithIdentity(s.Ctx, s, action.Identity, nodeID)
		err := node.ReEnableNAC(ctx)

		expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)
		assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)
		if !expectedErrorRaised {
			require.Equal(s.T, action.ExpectedError, "")
		}
	}
}

// DisableNAC will attempt to temporarily disable DefraDB's Node ACP system.
type DisableNAC struct {
	// NodeID may hold the ID (index) of the node we want to disable the node acp on.
	//
	// If a value is not provided, then will disable node acp on all nodes.
	NodeID immutable.Option[int]

	// The identity of a user that is authorized to disable the node acp.
	// The identity of the user that is disabling node acp, this user must be authorized
	// to disable node acp, otherwise error will be returned.
	Identity immutable.Option[state.Identity]

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// disableNAC will attempt to start the node access control system.
func disableNAC(
	s *state.State,
	action DisableNAC,
) {
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		ctx := getContextWithIdentity(s.Ctx, s, action.Identity, nodeID)
		err := node.DisableNAC(ctx)

		expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)
		assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)
		if !expectedErrorRaised {
			require.Equal(s.T, action.ExpectedError, "")
		}
	}
}

// AddNACActorRelationship will attempt to create a new relationship for node access with an actor.
type AddNACActorRelationship struct {
	// NodeID may hold the ID (index) of the node we want to add actor relationship on.
	//
	// If a value is not provided the relationship will be added in all nodes.
	NodeID immutable.Option[int]

	// The name of the relation to set for target actor (MUST be defined in the uploaded nac policy).
	//
	// This is a required field.
	Relation string

	// The target public identity, i.e. the identity of the actor to create a relationship with.
	//
	// This is a required field. To test the invalid usage of not having this arg, use NoIdentity() or leave default.
	TargetIdentity immutable.Option[state.Identity]

	// The requestor identity, i.e. identity of the actor creating the relationship.
	// Note: This identity must either own or have managing access defined in the policy.
	//
	// This is a required field. To test the invalid usage of not having this arg, use NoIdentity() or leave default.
	RequestorIdentity immutable.Option[state.Identity]

	// Result returns true if it was a no-op due to existing before, and false if a new relationship was made.
	ExpectedExistence bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

func addNACActorRelationship(
	s *state.State,
	action AddNACActorRelationship,
) {
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		addActorRelationshipResult, err := node.AddNACActorRelationship(
			getContextWithIdentity(s.Ctx, s, action.RequestorIdentity, nodeID),
			action.Relation,
			getIdentityDID(s, action.TargetIdentity),
		)

		expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)
		assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(s.T, action.ExpectedError, "")
			require.Equal(s.T, action.ExpectedExistence, addActorRelationshipResult.ExistedAlready)
		}
	}
}

// DeleteNACActorRelationship will attempt to delete a relationship for node access with an actor.
type DeleteNACActorRelationship struct {
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
	TargetIdentity immutable.Option[state.Identity]

	// The requestor identity, i.e. identity of the actor deleting the relationship.
	// Note: This identity must either own or have managing access defined in the policy.
	//
	// This is a required field. To test the invalid usage of not having this arg, use NoIdentity() or leave default.
	RequestorIdentity immutable.Option[state.Identity]

	// Result returns true if the relationship record was expected to be found and deleted,
	// and returns false if no matching relationship record was found (no-op).
	ExpectedRecordFound bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

func deleteNACActorRelationship(
	s *state.State,
	action DeleteNACActorRelationship,
) {
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		deleteActorRelationshipResult, err := node.DeleteNACActorRelationship(
			getContextWithIdentity(s.Ctx, s, action.RequestorIdentity, nodeID),
			action.Relation,
			getIdentityDID(s, action.TargetIdentity),
		)

		expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)
		assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(s.T, action.ExpectedError, "")
			require.Equal(s.T, action.ExpectedRecordFound, deleteActorRelationshipResult.RecordFound)
		}
	}
}

// GetNACStatus returns the current status of the target node(s) if all goes well,
// otherwise returns an error.
type GetNACStatus struct {
	// NodeID may hold the ID (index) of the node we want to check node acp status on.
	//
	// If a value is not provided, then will check node acp status on all nodes.
	NodeID immutable.Option[int]

	// The identity of the user that is making this request.
	Identity immutable.Option[state.Identity]

	// ExpectedStatus returns the current status of the NAC system.
	ExpectedStatus client.NACStatus

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

func getNACStatus(
	s *state.State,
	action GetNACStatus,
) {
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		statusNACResult, err := node.GetNACStatus(
			getContextWithIdentity(s.Ctx, s, action.Identity, nodeID),
		)

		expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)
		assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(s.T, action.ExpectedError, "")
			require.Equal(s.T, action.ExpectedStatus.String(), statusNACResult.Status)
		}
	}
}
