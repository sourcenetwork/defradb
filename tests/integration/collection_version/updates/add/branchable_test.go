// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package add

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestBranchableCollection_AddNewField_ShouldUpdateCollectionDefinition(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User @branchable {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/User/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetCollectionName("User"),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "User",
						IsBranchable:   true,
						IsMaterialized: true,
						IsActive:       true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "bafyreibhpgygzsmki22sql5ejzcojrrxbc5iuhpydhdzxul5w2znc7zrgu",
						}),
						Fields: []client.CollectionFieldDescription{
							{
								Name: request.DocIDFieldName,
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "email",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestBranchableCollection_AddNewFieldWithMultipleDocs_ShouldAddField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User @branchable {
						name: String
					}
				`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				DocMap: map[string]any{
					"name": "Islam",
				},
			},
			&action.PatchCollection{
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/User/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			&action.AddDoc{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
