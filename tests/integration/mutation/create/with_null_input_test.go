// Copyright 2024 Democratized Data Foundation
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
)

func TestMutationCreate_WithNullEncrypt_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation, with null encrypt",
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
					create_Users(encrypt: null, input: {name: "Bob"}) {
						name
					}
				}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
						{
							"name": "Bob",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithNullInput_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation, with null input",
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
					create_Users(input: null) {
						name
					}
				}`,
				Results: map[string]any{
					"create_Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithNullInputEntry_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation, with null input entry returns error",
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
					create_Users(input: [null]) {
						name
					}
				}`,
				ExpectedError: "Expected \"UsersMutationInputArg!\", found null.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithNullEncryptFields_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple create mutation, with null encryptFields",
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
					create_Users(encryptFields: null, input: {name: "Bob"}) {
						name
					}
				}`,
				Results: map[string]any{
					"create_Users": []map[string]any{
						{
							"name": "Bob",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
