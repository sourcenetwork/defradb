// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package remove

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesRemoveCollectionNameErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove collection name",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "remove", "path": "/Users/Name" }
					]
				`,
				ExpectedError: "schema name can't be empty",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesRemoveSchemaRootErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove schema root",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "remove", "path": "/Users/Root" }
					]
				`,
				ExpectedError: "SchemaRoot does not match existing",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesRemoveSchemaVersionIDErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove schema version id",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// This should do nothing
				Patch: `
					[
						{ "op": "remove", "path": "/Users/VersionID" }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesRemoveSchemaNameErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove schema name",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "remove", "path": "/Users/Name" }
					]
				`,
				ExpectedError: "schema name can't be empty",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
