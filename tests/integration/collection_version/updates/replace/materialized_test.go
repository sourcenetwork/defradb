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
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafyreiezxal4wrjp2fn6x5pf3kecliun72ky5tvb4deql2j376bmdknuh4/IsMaterialized",
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
			testUtils.CreateView{
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
			testUtils.CreateDoc{
				// Create John when the view is materialized
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.PatchCollection{
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
			testUtils.CreateDoc{
				// Create Fred when the view is not materialized, noting that there is no `RefreshView`
				// call after this action, meaning that if the view was still materialized Fred would not
				// be returned by the query.
				DocMap: map[string]any{
					"name": "Fred",
				},
			},
			testUtils.GetCollections{
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
			testUtils.Request{
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
