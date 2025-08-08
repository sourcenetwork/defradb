// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package fields

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesRemoveFieldErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove field",
		Actions: []any{
			&action.AddSchema{
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
						{ "op": "remove", "path": "/Users/Fields/2" }
					]
				`,
				ExpectedError: "deleting an existing field is not supported. Name: name",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesRemoveAllFieldsErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove all fields",
		Actions: []any{
			&action.AddSchema{
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
						{ "op": "remove", "path": "/Users/Fields" }
					]
				`,
				ExpectedError: "deleting an existing field is not supported",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesRemoveFieldNameErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove field name",
		Actions: []any{
			&action.AddSchema{
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
						{ "op": "remove", "path": "/Users/Fields/2/Name" }
					]
				`,
				ExpectedError: "deleting an existing field is not supported. Name: name",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesRemoveFieldKindErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove field kind",
		Actions: []any{
			&action.AddSchema{
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
						{ "op": "remove", "path": "/Users/Fields/2/Kind" }
					]
				`,
				ExpectedError: "mutating an existing field is not supported. ProposedName: ",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesRemoveFieldTypErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove field Typ",
		Actions: []any{
			&action.AddSchema{
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
						{ "op": "remove", "path": "/Users/Fields/2/Typ" }
					]
				`,
				ExpectedError: "mutating an existing field is not supported. ProposedName: name",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
