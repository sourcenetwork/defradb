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

func TestSchemaUpdatesTestFieldNameErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, test field name passes",
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
						{ "op": "test", "path": "/Users/Fields/1/name", "value": "Email" }
					]
				`,
				ExpectedError: "testing value /Users/Fields/1/name failed: test failed",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesTestFieldNamePasses(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, test field name passes",
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
						{ "op": "test", "path": "/Users/Fields/1/Name", "value": "name" }
					]
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesTestFieldErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, test field fails",
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
						{ "op": "test", "path": "/Users/Fields/1", "value": {"Name": "name", "Kind": 11} }
					]
				`,
				ExpectedError: "testing value /Users/Fields/1 failed: test failed",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesTestFieldPasses(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, test field passes",
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
						{ "op": "test", "path": "/Users/Fields/1", "value": {
							"Name": "name", "Kind": 11, "Typ":1
						} }
					]
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesTestFieldPasses_UsingFieldNameAsIndex(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, test field passes",
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
						{ "op": "test", "path": "/Users/Fields/name", "value": {
							"Kind": 11, "Typ":1
						} }
					]
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesTestFieldPasses_TargettingKindUsingFieldNameAsIndex(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, test field passes",
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
						{ "op": "test", "path": "/Users/Fields/name/Kind", "value": 11 }
					]
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
