// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package replace

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesReplaceCollectionErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, replace collection",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// Replace Users with Book
				Patch: `
					[
						{
							"op": "replace", "path": "/Users", "value": {
								"Name": "Book",
								"Schema": {
									"Name": "Book",
									"Fields": [
										{"Name": "name", "Kind": 11}
									]
								} 
							}
						}
					]
				`,
				// WARNING: An error is still expected if/when we allow the adding of collections, as this also
				// implies that the "Users" collection is to be deleted.  Only once we support the adding *and*
				// removal of collections should this not error.
				ExpectedError: "unknown collection, adding collections via patch is not supported. Name: Book",
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

/* WIP
func TestSchemaUpdatesReplaceCollectionNameWithExistingDoesNotChangeVersionID(t *testing.T) {
	schemaVersionID := "bafkreicg3xcpjlt3ecguykpcjrdx5ogi4n7cq2fultyr6vippqdxnrny3u"

	test := testUtils.TestCase{
		Description: "Test schema update, replacing collection name with self does not change version ID",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
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
			testUtils.SchemaPatch{
				// This patch essentially does nothing, replacing the current value with the current value
				Patch: `
					[
						{ "op": "replace", "path": "/Users/Name", "value": "Users" }
					]
				`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"name": "Johnnn"
				}`,
			},
			testUtils.Request{
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
*/
