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

func TestSchemaUpdatesAddFieldSimple(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
						Email
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaUpdatesAddFieldSimpleErrorsAddingToUnknownCollection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add to unknown collection fails",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Authors/Schema/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
				ExpectedError: "add operation does not apply: doc is missing path",
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaUpdatesAddFieldMultipleInPatch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add multiple fields in single patch",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Email", "Kind": 11} },
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "City", "Kind": 11} }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
						Email
						City
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaUpdatesAddFieldMultiplePatches(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add multiple patches",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "City", "Kind": 11} }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
						Email
						City
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaUpdatesAddFieldSimpleWithoutName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field without name",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Kind": 11} }
					]
				`,
				ExpectedError: "Names must match /^[_a-zA-Z][_a-zA-Z0-9]*$/ but \"\" does not.",
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaUpdatesAddFieldMultipleInPatchPartialSuccess(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add multiple fields in single patch with rollback",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// Email field is valid, City field has invalid kind
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Email", "Kind": 11} },
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "City", "Kind": 111} }
					]
				`,
				ExpectedError: "no type found for given name. Type: 111",
			},
			testUtils.Request{
				// Email does not exist as the commit failed
				Request: `query {
					Users {
						Name
						Email
					}
				}`,
				ExpectedError: "Cannot query field \"Email\" on type \"Users\"",
			},
			testUtils.Request{
				// Original schema is preserved
				Request: `query {
					Users {
						Name
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaUpdatesAddFieldSimpleDuplicateOfExistingField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field that already exists",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Name", "Kind": 11} }
					]
				`,
				ExpectedError: "duplicate field. Name: Name",
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaUpdatesAddFieldSimpleDuplicateField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add duplicate fields",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Email", "Kind": 11} },
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
				ExpectedError: "duplicate field. Name: Email",
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaUpdatesAddFieldWithExplicitIDErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field that already exists",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"ID": 2, "Name": "Email", "Kind": 11} }
					]
				`,
				ExpectedError: "explicitly setting a field ID value is not supported. Field: Email, ID: 2",
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
