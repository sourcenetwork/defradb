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

func TestSchemaUpdatesCopyFieldIntrospectionWithRemoveIDAndReplaceName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, copy and replace field with gql introspection",
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
						{ "op": "copy", "from": "/Users/Fields/1", "path": "/Users/Fields/2" },
						{ "op": "remove", "path": "/Users/Fields/2/ID" },
						{ "op": "replace", "path": "/Users/Fields/2/Name", "value": "fax" }
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
						).Append(
							introspectionUtils.Field{
								"name": "fax",
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
	testUtils.ExecuteTestCase(t, test)
}
