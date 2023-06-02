// Copyright 2022 Democratized Data Foundation
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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	simpleTests "github.com/sourcenetwork/defradb/tests/integration/mutation/simple"
)

func TestMutationCreateSimpleErrorsGivenNonExistantField(t *testing.T) {
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
			testUtils.Request{
				Request: `mutation {
							create_Users(data: "{\"name\": \"John\",\"fieldDoesNotExist\": 27}") {
								_key
							}
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

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestMutationCreateSimple(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple create mutation",
		Request: `mutation {
					create_User(data: "{\"name\": \"John\",\"age\": 27,\"points\": 42.1,\"verified\": true}") {
						_key
						name
						age
					}
				}`,
		Results: []map[string]any{
			{
				"_key": "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
				"age":  uint64(27),
				"name": "John",
			},
		},
	}

	simpleTests.ExecuteTestCase(t, test)
}

func TestMutationCreateSimpleDoesNotCreateDocGivenDuplicate(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple create mutation where document already exists.",
		Request: `mutation {
					create_User(data: "{\"name\": \"John\",\"age\": 27}") {
						_key
						name
						age
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"age": 27
				}`,
			},
		},
		ExpectedError: "a document with the given dockey already exists. DocKey: ",
	}

	simpleTests.ExecuteTestCase(t, test)
}

func TestMutationCreateSimpleDoesNotCreateDocEmptyData(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple create mutation with empty data param.",
		Request: `mutation {
					create_User(data: "") {
						_key
						name
						age
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"age": 27
				}`,
			},
		},
		ExpectedError: "given data payload is empty",
	}

	simpleTests.ExecuteTestCase(t, test)
}
