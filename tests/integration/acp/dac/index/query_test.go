// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_dac_index

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACPWithIndex_UponQueryingPrivateDocWithoutIdentity_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, querying private doc without identity should not fetch",
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   userPolicy,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String @index
						age: Int
					}
				`,

				Replace: map[string]testUtils.ReplaceType{
					"Policy0": testUtils.NewPolicyIndex(0),
				},
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name": "Shahzad"
					}
				`,
			},
			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),
				Doc: `
					{
						"name": "Islam"
					}
				`,
			},
			testUtils.Request{
				Request: `
					query  {
						Users {
							name
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Shahzad"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACPWithIndex_UponQueryingPrivateDocWithIdentity_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, querying private doc with identity should  fetch",
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   userPolicy,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String @index
						age: Int
					}
				`,

				Replace: map[string]testUtils.ReplaceType{
					"Policy0": testUtils.NewPolicyIndex(0),
				},
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name": "Shahzad"
					}
				`,
			},
			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),
				Doc: `
					{
						"name": "Islam"
					}
				`,
			},
			testUtils.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query  {
						Users {
							name
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Islam",
						},
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACPWithIndex_UponQueryingPrivateDocWithWrongIdentity_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, querying private doc with wrong identity should not fetch",
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   userPolicy,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String @index
						age: Int
					}
				`,

				Replace: map[string]testUtils.ReplaceType{
					"Policy0": testUtils.NewPolicyIndex(0),
				},
			},
			testUtils.CreateDoc{
				Doc: `
					{
						"name": "Shahzad"
					}
				`,
			},
			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),
				Doc: `
					{
						"name": "Islam"
					}
				`,
			},
			testUtils.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query  {
						Users {
							name
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
