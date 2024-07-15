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
	"math/rand"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/internal/db"
)

// AddPolicy will attempt to add the given policy using DefraDB's ACP system.
type AddPolicy struct {
	// NodeID may hold the ID (index) of the node we want to add policy to.
	//
	// If a value is not provided the policy will be added in all nodes.
	NodeID immutable.Option[int]

	// The raw policy string.
	Policy string

	// The policy creator identity, i.e. actor creating the policy.
	Identity immutable.Option[int]

	// The expected policyID generated based on the Policy loaded in to the ACP system.
	ExpectedPolicyID string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// addPolicyACP will attempt to add the given policy using DefraDB's ACP system.
func addPolicyACP(
	s *state,
	action AddPolicy,
) {
	// If we expect an error, then ExpectedPolicyID should be empty.
	if action.ExpectedError != "" && action.ExpectedPolicyID != "" {
		require.Fail(s.t, "Expected error should not have an expected policyID with it.", s.testCase.Description)
	}

	for _, node := range getNodes(action.NodeID, s.nodes) {
		identity := getIdentity(s, action.Identity)
		ctx := db.SetContextIdentity(s.ctx, identity)
		policyResult, err := node.AddPolicy(ctx, action.Policy)

		if err == nil {
			require.Equal(s.t, action.ExpectedError, "")
			require.Equal(s.t, action.ExpectedPolicyID, policyResult.PolicyID)
		}

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
	}
}

func getIdentity(s *state, index immutable.Option[int]) immutable.Option[acpIdentity.Identity] {
	if !index.HasValue() {
		return immutable.None[acpIdentity.Identity]()
	}

	if len(s.identities) <= index.Value() {
		// Generate the keys using the index as the seed so that multiple
		// runs yield the same private key.  This is important for stuff like
		// the change detector.
		source := rand.NewSource(int64(index.Value()))
		r := rand.New(source)

		privateKey, err := secp256k1.GeneratePrivateKeyFromRand(r)
		require.NoError(s.t, err)

		identity, err := acpIdentity.FromPrivateKey(privateKey)
		require.NoError(s.t, err)

		s.identities = append(s.identities, identity.Value())
		return identity
	} else {
		return immutable.Some[acpIdentity.Identity](s.identities[index.Value()])
	}
}
