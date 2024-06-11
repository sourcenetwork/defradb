// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package field_kinds

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationCreate_WithJSONFieldGivenValidJSON_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with JSON field given a valid JSON string.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
					create_Users(input: {name: "John", custom: "{\"tree\": \"maple\", \"age\": 250}"}) {
						_docID
						name
						custom
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-84ae4ef8-ca0c-5f32-bc85-cee97e731bc0",
						"custom": "{\"tree\":\"maple\",\"age\":250}",
						"name":   "John",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithJSONFieldGivenInvalidJSON_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with JSON field given a valid JSON string.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
					create_Users(input: {name: "John", custom: "{\"tree\": \"maple, \"age\": 250}"}) {
						_docID
						name
						custom
					}
				}`,
				ExpectedError: `Argument "input" has invalid value {name: "John", custom: "{\"tree\": \"maple, \"age\": 250}"}.
In field "custom": Expected type "JSON", found "{\"tree\": \"maple, \"age\": 250}".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithJSONFieldGivenSimpleString_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with JSON field given a valid JSON string.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
					create_Users(input: {name: "John", custom: "blah"}) {
						_docID
						name
						custom
					}
				}`,
				ExpectedError: `Argument "input" has invalid value {name: "John", custom: "blah"}.
In field "custom": Expected type "JSON", found "blah".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
