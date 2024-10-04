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
						_docID
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "bae-a948a3b2-3e89-5654-b0f0-71685a66b4d7",
							"custom": map[string]any{
								"tree": "maple",
								"age":  uint64(250),
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
						_docID
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "bae-90fd8b1b-bd11-56b5-a78c-2fb6f7b4dca0",
							"custom": []any{"maple", uint64(250)},
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
						_docID
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "bae-dd7c12f5-a7c5-55c6-8b35-ece853ae7f9e",
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
						_docID
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "bae-59731737-8793-5794-a9a5-0ed0ad696d5c",
							"custom": uint64(250),
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
						_docID
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "bae-608582c3-979e-5f34-80f8-a70fce875d05",
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
						_docID
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "bae-0c4b39cf-433c-5a9c-9bed-1e2796c35d14",
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
						_docID
						name
						custom
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "bae-f405f600-56d9-5de4-8d02-75fdced35e3b",
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
