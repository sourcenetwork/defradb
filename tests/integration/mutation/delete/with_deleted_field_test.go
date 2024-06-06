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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationDeletion_WithDeletedField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
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
						delete_User(docID: "bae-1ef746f8-821e-586f-99b2-4cb1fb9b782f") {
							_deleted
							_docID
						}
					}`,
				Results: []map[string]any{
					{
						"_deleted": true,
						"_docID":   "bae-1ef746f8-821e-586f-99b2-4cb1fb9b782f",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
