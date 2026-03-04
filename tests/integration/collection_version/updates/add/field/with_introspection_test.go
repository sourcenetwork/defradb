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

func TestCollectionVersionUpdatesAddFieldIntrospection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": 11} }
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
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddFieldIntrospectionDoesNotAmendGQLTypesGivenBadPatch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.PatchCollection{
				// The [Name] field is valid, but [Email] has an invalid [Kind].
				// [Name] should not be added to the GQL types.
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": 11} },
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 111} }
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
	testUtils.ExecuteTestCase(t, test)
}
