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
