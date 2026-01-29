// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package replace

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateReplaceIsMaterialized_GivenFalseAndCollection_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
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
			&action.AddSchema{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			&action.CreateView{
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
			&action.CreateDoc{
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
			&action.CreateDoc{
				// Create Fred when the view is not materialized, noting that there is no `RefreshView`
				// call after this action, meaning that if the view was still materialized Fred would not
				// be returned by the query.
				DocMap: map[string]any{
					"name": "Fred",
				},
			},
			&action.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					Name: immutable.Some("UserView"),
				},
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
