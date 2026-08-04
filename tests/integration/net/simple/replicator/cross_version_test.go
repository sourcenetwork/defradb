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

package replicator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// crossVersion is the older released version exercised against the current
// build. v1.0.0 is the only post-1.0 release, so it is the only pair that
// matters today.
const crossVersion = "v1.0.0"

// waitForSyncOn polls the node until a User named name appears.
//
// WaitForSync reads a node's event bus, which an external node runs in
// another process and cannot expose, so we poll a query instead.
func waitForSyncOn(nodeID int, name string) *action.RunFunc {
	return action.NewRunFuncWithState(func(s *state.State) {
		require.Eventually(s.T, func() bool {
			result := s.Nodes[nodeID].ExecRequest(s.Ctx, `query { User { name } }`)
			if len(result.GQL.Errors) > 0 {
				return false
			}
			return hasUserWithName(result.GQL.Data, name)
		}, 30*time.Second, 200*time.Millisecond, "document %q did not sync to node %d in time", name, nodeID)
	})
}

// hasUserWithName reports whether the query result holds a User named name.
// The native client returns []map[string]any and the HTTP client returns
// []any, so both are handled.
func hasUserWithName(data any, name string) bool {
	m, ok := data.(map[string]any)
	if !ok {
		return false
	}
	switch users := m["User"].(type) {
	case []any:
		for _, u := range users {
			if um, ok := u.(map[string]any); ok && um["name"] == name {
				return true
			}
		}
	case []map[string]any:
		for _, u := range users {
			if u["name"] == name {
				return true
			}
		}
	}
	return false
}

// A doc created on the current build replicates to a v1.0.0 node.
// Skips when v1.0.0 has no release asset for this platform.
func TestP2PCrossVersion_HeadToV1_DocSyncs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(), // node 0 = HEAD
			testUtils.RandomNetworkingConfig().WithVersion(crossVersion), // node 1 = v1.0.0
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc:    `{"name": "Alice"}`,
			},
			waitForSyncOn(1, "Alice"),
			&action.Request{
				NodeID:  immutable.Some(1),
				Request: `query { User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Alice"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A doc created on a v1.0.0 node replicates to the current build.
// Skips when v1.0.0 has no release asset for this platform.
func TestP2PCrossVersion_V1ToHead_DocSyncs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(), // node 0 = HEAD
			testUtils.RandomNetworkingConfig().WithVersion(crossVersion), // node 1 = v1.0.0
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			&action.AddDoc{
				NodeID: immutable.Some(1),
				Doc:    `{"name": "Bob"}`,
			},
			waitForSyncOn(0, "Bob"),
			&action.Request{
				NodeID:  immutable.Some(0),
				Request: `query { User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{
						{"name": "Bob"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
