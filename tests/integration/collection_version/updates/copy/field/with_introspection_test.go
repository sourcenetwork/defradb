// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package field

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	introspectionUtils "github.com/sourcenetwork/defradb/tests/integration/collection_version"
)

func TestCollectionVersionUpdatesCopyFieldIntrospectionWithRemoveIDAndReplaceName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "copy", "from": "/Users/Fields/1", "path": "/Users/Fields/2" },
						{ "op": "replace", "path": "/Users/Fields/2/Name", "value": "fax" },
						{ "op": "remove", "path": "/Users/Fields/2/FieldID" }
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
