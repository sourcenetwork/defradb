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

package truncate

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestTruncateCollectionParallel_DeletesAllPreviouslyExistingDocuments(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			// Create the first docs before parallel, so that deletion can be verified
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "Islam",
				},
			},
			&action.Parallel{
				Children: []action.Action{
					&action.AddDoc{
						DoNotWaitForEvent: true,
						DocMap: map[string]any{
							"name": "Fred",
						},
					},
					&action.AddDoc{
						DoNotWaitForEvent: true,
						DocMap: map[string]any{
							"name": "Shahzad",
						},
					},
					&action.AddDoc{
						DoNotWaitForEvent: true,
						DocMap: map[string]any{
							"name": "Chris",
						},
					},
					&action.Request{
						// Filter by a name that doesn't exist, this will scan the entire collection, and
						// because it does not exist, we don't have to worry about a complicated result-assert
						//
						// We just add this action in to try and protect against any errors/panics/etc that might
						// come about from it - it should never have an impact on the end result.
						Request: `query {
							Users(filter: {name: {_eq: "Keenan"}}) {
								name
							}
						}`,
						Results: map[string]any{
							"Users": []map[string]any{},
						},
					},
					&action.Truncate{},
				},
			},
			// After the concurrent action has completed, make sure that `John` and `Islam` dont exist.
			//
			// Other documents might exist depending on the order of execution, and this test does not care,
			// and cannot guarantee, which ones will exist when this action executes.
			&action.Request{
				Request: `query {
					Users(filter: {_or: [{name: {_eq: "John"}},{name: {_eq: "Islam"}}]}) {
						name
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
