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

package add

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationAdd_WithNullEncrypt_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.Request{
				Request: `mutation {
					add_Users(encrypt: null, input: {name: "Bob"}) {
						name
					}
				}`,
				Results: map[string]any{
					"add_Users": []map[string]any{
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

func TestMutationAdd_WithNullInput_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.Request{
				Request: `mutation {
					add_Users(input: null) {
						name
					}
				}`,
				Results: map[string]any{
					"add_Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd_WithNullInputEntry_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.Request{
				Request: `mutation {
					add_Users(input: [null]) {
						name
					}
				}`,
				ExpectedError: "Expected \"UsersMutationInputArg!\", found null.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd_WithNullEncryptFields_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.Request{
				Request: `mutation {
					add_Users(encryptFields: null, input: {name: "Bob"}) {
						name
					}
				}`,
				Results: map[string]any{
					"add_Users": []map[string]any{
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
