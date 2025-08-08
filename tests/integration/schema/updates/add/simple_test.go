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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesAddSimpleErrorsAddingSchema(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add schema fails",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/-", "value": {"Name": "books"} }
					]
				`,
				ExpectedError: "adding collections via patch is not supported. Name: books",
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
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

// This test documents unwanted behaviour, it should fail but it does not, likely due to the `Sources` json
// deserialization magic.  This bug is documented by https://github.com/sourcenetwork/defradb/issues/3795
func TestSchemaUpdatesAddSimpleErrorsAddingSchemaProp(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add schema property fails",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/-", "value": {"Foo": "Bar"} }
					]
				`,
				ExpectedError: "",
				// Below is the desired error
				//ExpectedError: `json: unknown field "-"`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddSimpleErrorsAddingUnsupportedCollectionProp(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add to unsupported collection prop",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Foo/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
				ExpectedError: "add operation does not apply: doc is missing path",
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
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

func TestSchemaUpdatesAddSimpleErrorsAddingUnsupportedSchemaProp(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add to unsupported schema prop",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Foo/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
				ExpectedError: "add operation does not apply: doc is missing path",
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
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
