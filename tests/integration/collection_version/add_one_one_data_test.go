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

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestAddOneToOne_Input_PrimaryObject(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						author: Author
					}

					type Author {
						name: String
						wrote: Book @primary
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "AuthorMutationInputArg") {
							name
							inputFields {
								name
								type {
									name
									ofType {
										name
										kind
									}
								}
							}
						}
					}
				`,
				ContainsData: map[string]any{
					"__type": map[string]any{
						"name": "AuthorMutationInputArg",
						"inputFields": []any{
							map[string]any{
								"name": "name",
								"type": map[string]any{
									"name":   "String",
									"ofType": nil,
								},
							},
							map[string]any{
								"name": "wrote",
								"type": map[string]any{
									"name":   "ID",
									"ofType": nil,
								},
							},
							map[string]any{
								"name": "wrote",
								"type": map[string]any{
									"name":   "ID",
									"ofType": nil,
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAddOneToOne_Input_SecondaryObject(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						author: Author
					}

					type Author {
						name: String
						wrote: Book @primary
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "BookMutationInputArg") {
							name
							inputFields {
								name
								type {
									name
									ofType {
										name
										kind
									}
								}
							}
						}
					}
				`,
				ContainsData: map[string]any{
					"__type": map[string]any{
						"name": "BookMutationInputArg",
						// Note: the secondary relation fields should not be here!
						"inputFields": []any{
							map[string]any{
								"name": "name",
								"type": map[string]any{
									"name":   "String",
									"ofType": nil,
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
