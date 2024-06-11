// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package create

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationCreate_GivenNonExistantField_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation with non existant field",
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GQL mutation will return a different error
			// when field types do not match
			testUtils.CollectionNamedMutationType,
			testUtils.CollectionSaveMutationType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"fieldDoesNotExist": 27
				}`,
				ExpectedError: "The given field does not exist. Name: fieldDoesNotExist",
			},
			testUtils.Request{
				// Ensure that no documents have been written.
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						Users {
							_docID
							name
							age
						}
					}
				`,
				Results: []map[string]any{
					{
						"_docID": "bae-8c89a573-c287-5d8c-8ba6-c47c814c594d",
						"name":   "John",
						"age":    int64(27),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_GivenDuplicate_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation where document already exists.",
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// Collection.Save would treat the second create as an update, and so
			// is excluded from this test.
			testUtils.CollectionNamedMutationType,
			testUtils.GQLRequestMutationType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
				ExpectedError: "a document with the given ID already exists",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_GivenEmptyInput(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation with empty input param.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
					create_Users(input: {}) {
						_docID
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": "bae-332de69b-47da-5175-863f-2480107f4884",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
