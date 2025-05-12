// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package delete

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationDeletion_WithDeletedField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.Request{
				Request: `mutation {
						delete_User(docID: "bae-0879efe9-8717-5e4c-a77f-c81a453dc952") {
							_deleted
							_docID
						}
					}`,
				Results: map[string]any{
					"delete_User": []map[string]any{
						{
							"_deleted": true,
							"_docID":   "bae-0879efe9-8717-5e4c-a77f-c81a453dc952",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
