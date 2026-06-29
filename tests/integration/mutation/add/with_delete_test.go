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

package add

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Re-adding the same payload after delete collides with the tombstoned
// genesis docID and must error.
func TestMutationAdd_GivenPreviouslyDeletedDoc_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			testUtils.DeleteDoc{
				DocID: 0,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
				ExpectedError: "a document with the given ID has been deleted",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd_GivenPreviouslyDeletedDoc_DocRemainsDeleted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			testUtils.DeleteDoc{
				DocID: 0,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
				ExpectedError: "a document with the given ID has been deleted",
			},
			&action.Request{
				Request: `
					query {
						Users {
							_docID
							name
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
