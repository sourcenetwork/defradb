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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColDescrUpdateReplaceQuerySourceQuery(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Books {
						name: String
					}
				`,
			},
			testUtils.CreateView{
				// Create the view on the `Books` collection
				Query: `
					Books {
						name
					}
				`,
				SDL: `
					type View @materialized(if: false) {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			testUtils.PatchCollection{
				// Patch the view query definition so that it now queries the `Users` collection
				Patch: `
					[
						{ "op": "replace", "path": "/3/Sources/0/Query", "value": {"Name": "Users", "Fields":[{"Name":"name"}]} }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					View {
						name
					}
				}`,
				// If the view was still querying `Books` there would be no results
				Results: map[string]any{
					"View": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdateReplaceQuerySourceQueryName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Books {
						name: String
					}
				`,
			},
			testUtils.CreateView{
				// Create the view on the `Books` collection
				Query: `
					Books {
						name
					}
				`,
				SDL: `
					type View @materialized(if: false) {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			testUtils.PatchCollection{
				// Patch the view query definition so that it now queries the `Users` collection
				Patch: `
					[
						{ "op": "replace", "path": "/3/Sources/0/Query/Name", "value": "Users" }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					View {
						name
					}
				}`,
				// If the view was still querying `Books` there would be no results
				Results: map[string]any{
					"View": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
