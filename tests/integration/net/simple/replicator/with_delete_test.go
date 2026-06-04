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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestP2POneToOneReplicatorDeletesDocAddedBeforeReplicatorConfig(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
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
				// This document is added in first node before the replicator is set up.
				// Updates should be synced across nodes.
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.DeleteDoc{
				// Delete John from the first node only, and allow the value to sync
				NodeID: immutable.Some(0),
				DocID:  0,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					Users(showDeleted: true) {
						_deleted
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_deleted": true,
							"Name":     "John",
							"Age":      int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2POneToOneReplicatorDeletesDocAddedBeforeReplicatorConfigWithNodesInversed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
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
				// This document is added in second node before the replicator is set up.
				// Updates should be synced across nodes.
				NodeID: immutable.Some(1),
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.DeleteDoc{
				// Delete John from the first node only, and allow the value to sync
				NodeID: immutable.Some(1),
				DocID:  0,
			},
			testUtils.WaitForSync{},
			&action.Request{
				Request: `query {
					Users(showDeleted: true) {
						_deleted
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_deleted": true,
							"Name":     "John",
							"Age":      int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
