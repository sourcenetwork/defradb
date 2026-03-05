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

package peer_replicator_test

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestP2PPeerReplicatorWithAdd(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 2,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Shahzad",
					"Age": 3000
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(3000),
						},
						{
							"Age": int64(21),
						},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(21),
						},
					},
				},
			},
			&action.Request{
				NodeID: immutable.Some(2),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Age": int64(3000),
						},
						{
							"Age": int64(21),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
