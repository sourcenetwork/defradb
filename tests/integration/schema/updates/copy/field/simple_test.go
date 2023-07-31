// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package field

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesCopyFieldErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, copy field",
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
						{ "op": "copy", "from": "/Users/Schema/Fields/1", "path": "/Users/Schema/Fields/2" }
					]
				`,
				ExpectedError: "duplicate field. Name: email",
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesCopyFieldWithRemoveIDAndReplaceName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, copy field, rename and remove IDs",
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
				// Here we esentially use Email as a template, copying it, clearing the ID, and renaming the
				// clone.
				Patch: `
					[
						{ "op": "copy", "from": "/Users/Schema/Fields/1", "path": "/Users/Schema/Fields/3" },
						{ "op": "remove", "path": "/Users/Schema/Fields/3/ID" },
						{ "op": "replace", "path": "/Users/Schema/Fields/3/Name", "value": "fax" }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						email
						fax
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// This is an odd test, but still a possibility and we should still cover it.
func TestSchemaUpdatesCopyFieldWithRemoveIDAndReplaceNameAndKindSubstitution(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, copy field, rename, re-type, and remove IDs",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// Here we esentially use Name as a template, copying it, clearing the ID, and renaming and
				// re-typing the clone.
				Patch: `
					[
						{ "op": "copy", "from": "/Users/Schema/Fields/1", "path": "/Users/Schema/Fields/2" },
						{ "op": "remove", "path": "/Users/Schema/Fields/2/ID" },
						{ "op": "replace", "path": "/Users/Schema/Fields/2/Name", "value": "age" },
						{ "op": "replace", "path": "/Users/Schema/Fields/2/Kind", "value": "Int" }
					]
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 3
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						age
					}
				}`,
				Results: []map[string]any{
					{
						"name": "John",
						// It is important to test this with data, to ensure the type has been substituted correctly
						"age": uint64(3),
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// This is an odd test, but still a possibility and we should still cover it.
func TestSchemaUpdatesCopyFieldWithRemoveIDAndReplaceNameAndInvalidKindSubstitution(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, copy field, rename, re-type to invalid, and remove ID",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// Here we esentially use Name as a template, copying it, clearing the ID, and renaming and
				// re-typing the clone.
				Patch: `
					[
						{ "op": "copy", "from": "/Users/Schema/Fields/1", "path": "/Users/Schema/Fields/2" },
						{ "op": "remove", "path": "/Users/Schema/Fields/2/ID" },
						{ "op": "replace", "path": "/Users/Schema/Fields/2/Name", "value": "Age" },
						{ "op": "replace", "path": "/Users/Schema/Fields/2/Kind", "value": "NotAValidKind" }
					]
				`,
				ExpectedError: "no type found for given name. Kind: NotAValidKind",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
