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

	"github.com/sourcenetwork/immutable"
)

func TestMutationCreate_WithJSONFieldGivenObjectValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with JSON field given an object value.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John", 
					"custom": {"tree": "maple", "age": 250}
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": map[string]any{
								"tree": "maple",
								"age":  float64(250),
							},
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithJSONFieldGivenListOfScalarsValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with JSON field given a list of scalars value.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John", 
					"custom": ["maple", 250]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": []any{"maple", float64(250)},
							"name":   "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithJSONFieldGivenListOfObjectsValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with JSON field given a list of objects value.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John", 
					"custom": [
						{"tree": "maple"}, 
						{"tree": "oak"}
					]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": []any{
								map[string]any{"tree": "maple"},
								map[string]any{"tree": "oak"},
							},
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithJSONFieldGivenIntValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with JSON field given a int value.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John", 
					"custom": 250
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": float64(250),
							"name":   "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithJSONFieldGivenStringValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with JSON field given a string value.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John", 
					"custom": "hello"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": "hello",
							"name":   "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithJSONFieldGivenBooleanValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with JSON field given a boolean value.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John", 
					"custom": true
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": true,
							"name":   "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithJSONFieldGivenNullValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with JSON field given a null value.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John", 
					"custom": null
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"custom": nil,
							"name":   "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test confirms that our JSON value encoding is determinstic.
func TestMutationCreate_WithDuplicateJSONField_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with duplicate JSON field errors.",
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// Save will not produce an error on duplicate
			// because it will just update the previous doc
			testUtils.GQLRequestMutationType,
			testUtils.CollectionNamedMutationType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						custom: JSON
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John", 
					"custom": {"one": 1, "two": 2, "three": [0, 1, 2]}
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John", 
					"custom": {"three": [0, 1, 2], "two": 2, "one": 1}
				}`,
				ExpectedError: `a document with the given ID already exists`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
