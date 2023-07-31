// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesTestCollectionNameErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, test collection name",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "test", "path": "/Users/Name", "value": "Book" }
					]
				`,
				ExpectedError: "testing value /Users/Name failed: test failed",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesTestCollectionNamePasses(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, test collection name passes",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "test", "path": "/Users/Name", "value": "Users" }
					]
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

/* WIP
func TestSchemaUpdatesTestCollectionNameDoesNotChangeVersionID(t *testing.T) {
	schemaVersionID := "bafkreicg3xcpjlt3ecguykpcjrdx5ogi4n7cq2fultyr6vippqdxnrny3u"

	test := testUtils.TestCase{
		Description: "Test schema update, test collection name does not change version ID",
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
				Patch: `
					[
						{ "op": "test", "path": "/Users/Name", "value": "Users" }
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
	testUtils.ExecuteTestCase(t, test)
}
*/
