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

package test_acp_dac_index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACPWithIndex_UponQueryingPrivateDocWithoutIdentity_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   userPolicy,
			},
			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String @index
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name": "Shahzad"
					}
				`,
			},
			&action.AddDoc{
				Identity: testUtils.ClientIdentity(1),
				Doc: `
					{
						"name": "Islam"
					}
				`,
			},
			&action.Request{
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
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   userPolicy,
			},
			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String @index
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name": "Shahzad"
					}
				`,
			},
			&action.AddDoc{
				Identity: testUtils.ClientIdentity(1),
				Doc: `
					{
						"name": "Islam"
					}
				`,
			},
			&action.Request{
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
							"name": "Shahzad",
						},
						{
							"name": "Islam",
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
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   userPolicy,
			},
			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String @index
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `
					{
						"name": "Shahzad"
					}
				`,
			},
			&action.AddDoc{
				Identity: testUtils.ClientIdentity(1),
				Doc: `
					{
						"name": "Islam"
					}
				`,
			},
			&action.Request{
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
