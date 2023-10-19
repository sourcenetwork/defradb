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

func TestMutationDeletion_WithIDUnknownValue(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Deletion using id that doesn't exist, where the collection is empty.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
					delete_User(ids: ["bae-d7546ac1-c133-5853-b866-9b9f926fe7e5"]) {
						_docID
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithIDUnknownValueAndUnrelatedRecordInCollection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Deletion using id that doesn't exist, where the collection is non-empty.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "Not the id below"
				}`,
			},
			testUtils.Request{
				Request: `mutation {
					delete_User(ids: ["bae-d7546ac1-c133-5853-b866-9b9f926fe7e5"]) {
						_docID
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
