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

func TestP2PReplicatorUpdateWithNewFieldSyncsDocsToOlderCollectionVersionMultistep(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.PatchCollection{
				// Patch the collection on the node that we will update the doc on
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			&action.UpdateDoc{
				// Update the new field on the first node only, and allow the value to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Email": "imnotyourbuddyguy@source.ca"
				}`,
			},
			&action.UpdateDoc{
				// Update the existing field on the first node only, and allow the value to sync
				// We need to make sure any errors caused by the first update do not break the sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Shahzad"
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						Name
						Email
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":  "Shahzad",
							"Email": "imnotyourbuddyguy@source.ca",
						},
					},
				},
			},
			&action.Request{
				// The second update should still be received by the second node, updating Name
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Shahzad",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PReplicatorUpdateWithNewFieldSyncsDocsToOlderCollectionVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.PatchCollection{
				// Patch the collection on the node that we will directly update the doc on
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			&action.UpdateDoc{
				// Update the new field and existing field on the first node only, and allow the values to sync
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Shahzad",
					"Email": "imnotyourbuddyguy@source.ca"
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						Name
						Email
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name":  "Shahzad",
							"Email": "imnotyourbuddyguy@source.ca",
						},
					},
				},
			},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Shahzad",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
