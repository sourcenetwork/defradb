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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateReplaceQuerySourceQuery(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddCollection{
				SDL: `
					type Books {
						name: String
					}
				`,
			},
			&action.AddView{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			&action.PatchCollection{
				// Patch the view query definition so that it now queries the `Users` collection
				Patch: `
					[
						{
							"op": "replace",
							"path": "/View/Query/Query",
							"value": {"Name": "Users", "Fields":[{"Name":"name"}]}
						}
					]
				`,
			},
			&action.Request{
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

func TestColVersionUpdateReplaceQuerySourceQueryName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddCollection{
				SDL: `
					type Books {
						name: String
					}
				`,
			},
			&action.AddView{
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
			&action.AddDoc{
				Doc: `{
					"name": "Shahzad"
				}`,
			},
			&action.PatchCollection{
				// Patch the view query definition so that it now queries the `Users` collection
				Patch: `
					[
						{
							"op": "replace",
							"path": "/View/Query/Query/Name",
							"value": "Users"
						}
					]
				`,
			},
			&action.Request{
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
