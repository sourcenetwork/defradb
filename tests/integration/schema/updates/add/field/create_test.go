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
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						_key
						name
						email
					}
				}`,
				Results: []map[string]any{
					{
						"_key":  "bae-43deba43-f2bc-59f4-9056-fef661b22832",
						"Name":  "John",
						"Email": nil,
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
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
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Email", "Kind": 11} }
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
						_key
						name
						email
					}
				}`,
				Results: []map[string]any{
					{
						"_key":  "bae-43deba43-f2bc-59f4-9056-fef661b22832",
						"name":  "John",
						"email": nil,
					},
					{
						"_key":  "bae-68926881-2eed-519b-b4eb-883b4a6624a6",
						"name":  "Shahzad",
						"email": "sqlizded@yahoo.ca",
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
