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
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"
)

// AddPolicy will attempt to add the given policy using defraDB's ACP module.
type AddPolicy struct {
	// NodeID may hold the ID (index) of the node we want to add policy to.
	//
	// If a value is not provided the policy will be added in all nodes.
	NodeID immutable.Option[int]

	// The raw policy string.
	Policy string

	// The policy creator, i.e. actor creating the policy.
	Creator string

	// If true then assumes YAML format for the policy, otherwise assumes JSON format.
	IsYAML bool

	// The expected policyID generated based on the Policy loaded in to acp module.
	ExpectedPolicyID string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// addPolicyACP will attempt to add the given policy using defraDB's ACP module.
func addPolicyACP(
	s *state,
	action AddPolicy,
) {
	if s.dbt == badgerIMType || s.dbt == defraIMType {
		s.t.Skip("ACP test (adding policy) is not supported on in-memory database type(s)")
	}

	// If we expect an error, then ExpectedPolicyID should be empty.
	if action.ExpectedError != "" && action.ExpectedPolicyID != "" {
		require.Fail(s.t, "Expected error should not have an expected policyID with it.", s.testCase.Description)
	}

	// for nodeID, node := range getNodes(action.NodeID, s.nodes) {
	for _, node := range getNodes(action.NodeID, s.nodes) {
		if !node.ACPModule().HasValue() {
			require.Fail(s.t, "Failed to add policy because ACP module was not found", s.testCase.Description)
		}

		policyID, err := node.ACPModule().Value().AddPolicy(
			s.ctx,
			action.Policy,
			action.Creator,
			action.IsYAML,
		)

		if err == nil {
			require.Equal(s.t, action.ExpectedError, "")
			require.Equal(s.t, action.ExpectedPolicyID, policyID)
			// s.policyIDs[nodeID] = append(s.policyIDs[nodeID], policyID) // TODO-ACP: Remove
		}

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
	}
}
