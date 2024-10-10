// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCreateOneToOne_Input_PrimaryObject(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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

func TestCreateOneToOne_Input_SecondaryObject(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
