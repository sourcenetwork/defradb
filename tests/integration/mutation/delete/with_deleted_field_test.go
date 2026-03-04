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

package delete

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationDeletion_WithDeletedField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John"
				}`,
			},
			&action.Request{
				Request: `mutation {
						delete_User(docID: "bae-7f4197fe-c647-5cc6-91bb-5f32229fd4cd") {
							_deleted
							_docID
						}
					}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"_deleted": true,
							"_docID":   "bae-7f4197fe-c647-5cc6-91bb-5f32229fd4cd",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
