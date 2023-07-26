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
	introspectionUtils "github.com/sourcenetwork/defradb/tests/integration/schema"
)

func TestSchemaUpdatesAddFieldIntrospection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with gql introspection",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "name", "Kind": 11} }
					]
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "Users") {
							name
							fields {
								name
								type {
									name
									kind
								}
							}
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": map[string]any{
						"name": "Users",
						"fields": introspectionUtils.DefaultFields.Append(
							introspectionUtils.Field{
								"name": "name",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "String",
								},
							},
						).Tidy(),
					},
				},
			},
		},
	}
	testUtils.ExecuteTEMP(t, test)
}

func TestSchemaUpdatesAddFieldIntrospectionDoesNotAmendGQLTypesGivenBadPatch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add invalid field with gql introspection",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
				`,
			},
			testUtils.SchemaPatch{
				// The [Name] field is valid, but [Email] has an invalid [Kind].
				// [Name] should not be added to the GQL types.
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "name", "Kind": 11} },
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "email", "Kind": 111} }
					]
				`,
				ExpectedError: "no type found for given name. Type: 111",
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "Users") {
							name
							fields {
								name
								type {
									name
									kind
								}
							}
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": map[string]any{
						"name": "Users",
						// No fields have been added to the GQL [Users] type.
						"fields": introspectionUtils.DefaultFields.Tidy(),
					},
				},
			},
		},
	}
	testUtils.ExecuteTEMP(t, test)
}
