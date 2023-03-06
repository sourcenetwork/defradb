// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package move

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesMoveCollectionDoesNothing(t *testing.T) {
	schemaVersionID := "bafkreicg3xcpjlt3ecguykpcjrdx5ogi4n7cq2fultyr6vippqdxnrny3u"

	test := testUtils.TestCase{
		Description: "Test schema update, move collection",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"Name": "John"
				}`,
			},
			testUtils.SchemaPatch{
				// This just moves an object to a new key in a temporary dictionary, it doesn't actually do
				// anything
				Patch: `
					[
						{ "op": "move", "from": "/Users", "path": "/Books" }
					]
				`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Name": "Johnnn"
				}`,
			},
			testUtils.Request{
				// Assert that Users is still Users
				Request: `query {
					Users {
						Name
					}
				}`,
				Results: []map[string]any{
					{
						"Name": "Johnnn",
					},
				},
			},
			testUtils.Request{
				// Assert that the version ID remains the same
				Request: `query {
					commits (field: "C") {
						schemaVersionId
					}
				}`,
				Results: []map[string]any{
					{
						// Update commit
						"schemaVersionId": schemaVersionID,
					},
					{
						// Create commit
						"schemaVersionId": schemaVersionID,
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
