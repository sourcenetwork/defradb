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

// This test documents a bug, see:
// https://github.com/sourcenetwork/defradb/issues/1846
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
						delete_User(id: "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad") {
							_deleted
							_docID
						}
					}`,
				Results: []map[string]any{
					{
						// This should be true, as it has been deleted.
						"_deleted": false,
						"_docID":   "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
