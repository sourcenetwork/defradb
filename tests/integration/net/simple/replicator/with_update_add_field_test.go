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
	"github.com/sourcenetwork/defradb/tests/state"
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
			// The second update should still be received by the second node,
			// updating Name. That node may be on an older release, which can hold
			// the document without being able to report the commit that carried
			// it, so there is nothing for WaitForSync to observe. Poll instead.
			action.NewEventually(&action.Request{
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
			}),
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
			// The second node may be on an older release, which can hold this
			// document without being able to report the commit that carried it.
			// There is nothing for WaitForSync to observe in that case, so poll
			// the result instead.
			action.NewEventually(&action.Request{
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
			}),
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A push that fails is retried from the document's heads, which carry the id of the collection version
// they were authored against rather than the collection's root id.
func TestP2PReplicatorUpdateWithNewFieldWithTargetNodeTemporarilyOffline_SyncsUpdate(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDatabaseTypes: immutable.Some(
			[]state.DatabaseType{
				// This test only supports file type databases since it requires the ability to
				// stop and start a node without losing data.
				testUtils.BadgerFileType,
			},
		),
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
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.WaitForSync{},
			&action.PatchCollection{
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			testUtils.Close{
				NodeID: immutable.Some(1),
			},
			&action.UpdateDoc{
				// The push to the offline target fails, so this update can only arrive by retry.
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Fred"
				}`,
			},
			testUtils.Start{
				NodeID: immutable.Some(1),
			},
			testUtils.WaitForSync{},
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
							"Name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
