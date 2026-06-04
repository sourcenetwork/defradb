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

package replace

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateReplaceIsMaterialized_GivenFalseAndCollection_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafyreibhpgygzsmki22sql5ejzcojrrxbc5iuhpydhdzxul5w2znc7zrgu/IsMaterialized",
							"value": false
						}
					]
				`,
				ExpectedError: "non-materialized collections are not supported. Collection: User",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateReplaceIsMaterialized_GivenFalseAndView(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView {
						name: String
					}
				`,
			},
			&action.AddDoc{
				// Create John when the view is materialized
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/UserView/IsMaterialized",
							"value": false
						}
					]
				`,
			},
			&action.AddDoc{
				// Create Fred when the view is not materialized, noting that there is no `RefreshView`
				// call after this action, meaning that if the view was still materialized Fred would not
				// be returned by the query.
				DocMap: map[string]any{
					"name": "Fred",
				},
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetCollectionName("UserView"),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "UserView",
						IsMaterialized: false,
						IsActive:       true,
					},
				},
			},
			&action.Request{
				Request: `query {
					User {
						name
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
						},
						{
							"name": "Fred",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
