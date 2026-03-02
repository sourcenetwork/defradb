// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package peer_test

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// TestP2PUpdate_WithLWWConcurrentDifferentFields_BothFieldsPreserved tests that
// when two nodes concurrently update different LWW fields on the same document,
// both field values are preserved after sync.
func TestP2PUpdate_WithLWWConcurrentDifferentFields_BothFieldsPreserved(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.AddDocumentSubscription{
				NodeID: 0,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.AddDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Alice"
				}`,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"age": 42
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					Users {
						name
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Alice",
							"age":  int64(42),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
