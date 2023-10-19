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
						"_docID": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
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

func TestMutationCreate_GivenEmptyData_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation with empty data param.",
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
					create_Users(data: "") {
						_docID
					}
				}`,
				ExpectedError: "given data payload is empty",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
