// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package kind

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesAddFieldKind8(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind deprecated (8)",
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
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "foo", "Kind": 8} }
					]
				`,
				ExpectedError: "no type found for given name. Type: 8",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKind9(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind deprecated (9)",
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
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "foo", "Kind": 9} }
					]
				`,
				ExpectedError: "no type found for given name. Type: 9",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKind13(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind deprecated (13)",
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
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "foo", "Kind": 13} }
					]
				`,
				ExpectedError: "no type found for given name. Type: 13",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKind14(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind deprecated (14)",
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
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "foo", "Kind": 14} }
					]
				`,
				ExpectedError: "no type found for given name. Type: 14",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKind15(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind deprecated (15)",
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
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "foo", "Kind": 15} }
					]
				`,
				ExpectedError: "no type found for given name. Type: 15",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// This test is currently the first unsupported value, if it becomes supported
// please update this test to be the newly lowest unsupported value.
func TestSchemaUpdatesAddFieldKind22(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind unsupported (22)",
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
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "foo", "Kind": 22} }
					]
				`,
				ExpectedError: "no type found for given name. Type: 22",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// Tests a semi-random but hardcoded unsupported kind to try and protect against anything odd permitting
// high values.
func TestSchemaUpdatesAddFieldKind198(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind unsupported (198)",
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
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "foo", "Kind": 198} }
					]
				`,
				ExpectedError: "no type found for given name. Type: 198",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindInvalidSubstitution(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind unsupported (198)",
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
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "foo", "Kind": "InvalidKind"} }
					]
				`,
				ExpectedError: "no type found for given name. Kind: InvalidKind",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
