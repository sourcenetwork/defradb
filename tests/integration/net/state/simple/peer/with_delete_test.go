// Copyright 2022 Democratized Data Foundation
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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

// The parent-child distinction in these tests is as much documentation and test
// of the test system as of production.  See it as a santity check of sorts.
func TestP2PWithMultipleDocumentsSingleDelete(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				// Create John on all nodes
				Doc: `{
					"Name": "John",
					"Age": 43
				}`,
			},
			testUtils.CreateDoc{
				// Create Andy on all nodes
				Doc: `{
					"Name": "Andy",
					"Age": 74
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.DeleteDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users {
						_deleted
						Name
						Age
					}
				}`,
				Results: []map[string]any{
					{
						"_deleted": false,
						"Name":     "Andy",
						"Age":      uint64(74),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestP2PWithMultipleDocumentsSingleDeleteWithShowDeleted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				// Create John on all nodes
				Doc: `{
					"Name": "John",
					"Age": 43
				}`,
			},
			testUtils.CreateDoc{
				// Create Andy on all nodes
				Doc: `{
					"Name": "Andy",
					"Age": 74
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.DeleteDoc{
				NodeID: immutable.Some(0),
				DocID:  0,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				Request: `query {
					Users(showDeleted: true) {
						_deleted
						Name
						Age
					}
				}`,
				Results: []map[string]any{
					{
						"_deleted": false,
						"Name":     "Andy",
						"Age":      uint64(74),
					},
					{
						"_deleted": true,
						"Name":     "John",
						"Age":      uint64(43),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
