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
			testUtils.Request{
				Request: `mutation {
					create_Users(input: {name: "John", custom: {tree: "maple", age: 250}}) {
						_docID
						name
						custom
					}
				}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
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

func TestMutationCreate_WithJSONFieldGivenListValue_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Create mutation with JSON field given a list value.",
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
					create_Users(input: {name: "John", custom: ["maple", 250]}) {
						_docID
						name
						custom
					}
				}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
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
			testUtils.Request{
				Request: `mutation {
					create_Users(input: {name: "John", custom: 250}) {
						_docID
						name
						custom
					}
				}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
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
			testUtils.Request{
				Request: `mutation {
					create_Users(input: {name: "John", custom: "hello"}) {
						_docID
						name
						custom
					}
				}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
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
			testUtils.Request{
				Request: `mutation {
					create_Users(input: {name: "John", custom: true}) {
						_docID
						name
						custom
					}
				}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
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
			testUtils.Request{
				Request: `mutation {
					create_Users(input: {name: "John", custom: null}) {
						_docID
						name
						custom
					}
				}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
						{
							"_docID": "bae-09fc6d72-daf7-5a61-9523-73a9fac7ce13",
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
