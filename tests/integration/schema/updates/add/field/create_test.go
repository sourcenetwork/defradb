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

func TestSchemaUpdatesAddFieldWithCreate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with create",
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						_docID
						name
						email
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-6845cfdf-cb0f-56a3-be3a-b5a67be5fbdc",
						"name":   "John",
						"email":  nil,
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldWithCreateAfterSchemaUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with create after schema update",
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
			// We want to make sure that this works across database versions, so we tell
			// the change detector to split here.
			testUtils.SetupComplete{},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Shahzad",
					"email": "sqlizded@yahoo.ca"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				Results: []map[string]any{
					{
						"name":  "Shahzad",
						"email": "sqlizded@yahoo.ca",
					},
					{
						"name":  "John",
						"email": nil,
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
