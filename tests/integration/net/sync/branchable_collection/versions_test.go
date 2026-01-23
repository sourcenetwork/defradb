// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package branchable_collection

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestBranchableCollectionSync_WithBranchedVersionsAndDocs_ShouldSync(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User @branchable {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "Islam",
				},
			},
			testUtils.PatchCollection{
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/User/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name":  "Andy",
					"email": "andy@gmail.com",
				},
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				DocID:  1,
				Doc: `{
					"email": "islam@gmail.com"
				}`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(1),
				DocMap: map[string]any{
					"name": "Shahzad",
				},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(1),
				DocMap: map[string]any{
					"name": "Islam",
				},
			},
			testUtils.PatchCollection{
				NodeID: immutable.Some(1),
				Patch: `
					[
						{ "op": "add", "path": "/User/Fields/-", "value": {"Name": "score", "Kind": 4} }
					]
				`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(1),
				DocMap: map[string]any{
					"name":  "Fred",
					"score": 100,
				},
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(1),
				DocID:  1,
				Doc: `{
					"score": 80
				}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.SyncBranchableCollection{
				NodeID:       1,
				CollectionID: 0,
			},
			&action.SyncBranchableCollection{
				NodeID:       0,
				CollectionID: 0,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					User {
						name
						email
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name":  "Fred",
							"email": nil,
						},
						{
							"name":  "Shahzad",
							"email": nil,
						},
						{
							"name":  "John",
							"email": nil,
						},
						{
							"name":  "Islam",
							"email": "islam@gmail.com",
						},
						{
							"name":  "Andy",
							"email": "andy@gmail.com",
						},
					},
				},
			},
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
						score
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name":  "Fred",
							"score": 100,
						},
						{
							"name":  "Shahzad",
							"score": nil,
						},
						{
							"name":  "John",
							"score": nil,
						},
						{
							"name":  "Islam",
							"score": 80,
						},
						{
							"name":  "Andy",
							"score": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
