// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package txn_testing

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

// This test runs SetActiveCollectionVersion inside of a transaction, and illustrates that commiting
// the transaction results in the version being changed.
func TestTxn_SetActiveCollectionVersion_WithCommit_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			testUtils.SetActiveCollectionVersion{
				TransactionID: immutable.Some(1),
				VersionID:     "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
			},
			testUtils.TransactionCommit{
				TransactionID: 1,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				ExpectedError: `Cannot query field "email" on type "Users".`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// This test runs SetActiveCollectionVersion inside of a transaction, and illustrates that not commiting
// the transaction results in the version not yet being changed.
func TestTxn_SetActiveCollectionVersion_WithoutCommit_VersionNotChanged(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			testUtils.SetActiveCollectionVersion{
				TransactionID: immutable.Some(1),
				VersionID:     "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
			},

			&action.Request{
				TransactionID: immutable.Some(2),
				Request: `query {
					Users {
						name
						email
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSyncColVersion_WithInitialColVersion_CanBeActivatedAndQueried(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				NodeID: immutable.Some(0),
				// Note - at the time of writing, having two fields of different kinds is important
				// and an important bug did not surface when testing with a single field/kind.
				Schema: `
                    type Users {
                        name: String
                        age: Int
                    }
                `,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"{{.CollectionVersionID0}}"},
			},
			testUtils.WaitForSync{},

			testUtils.SetActiveCollectionVersion{
				NodeID:        immutable.Some(1),
				TransactionID: immutable.Some(0),
				VersionID:     "{{.CollectionVersionID0}}",
			},
			testUtils.TransactionCommit{
				TransactionID: 0,
			},
			&action.Request{
				TransactionID: immutable.Some(1),
				NodeID:        immutable.Some(1),
				Request: `mutation {
                    add_Users(input: {age: 31, name: "Bob"}) {
                        _docID
                    }
                }`,
				ExpectedError: "fggdsd",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
