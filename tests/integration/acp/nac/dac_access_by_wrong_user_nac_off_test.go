// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_nac

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_Disabled_WithDACEnabled_AccessByNonNodeOwner_OwnsTheDocument_CanAccess(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			// Temporarily disable to allow a non-node-owner to own some documents, and test the disabled nac case.
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},
			// Make a non-node-owner own a document.
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(2),
				Policy:   examplePolicy,
			},
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(2),
				Schema:   `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(2),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},

			// Document owner and no NAC so can access as DAC will be satisfied.
			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
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

func TestNAC_Disabled_WithDACEnabled_AccessByNonNodeOwner_DoesNotOwnTheDocument_CanNotAccess(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			// Temporarily disable to allow a non-node-owner to own some documents, and test the disabled nac case.
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},
			// Make a non-node-owner own a document.
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(2),
				Policy:   examplePolicy,
			},
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(2),
				Schema:   `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(2),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},

			// Not authorized user, so can't see the private documents due to DAC.
			// Note: There is no error here because it's blocked by DAC not NAC.
			&action.Request{
				Identity: testUtils.ClientIdentity(3),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_Disabled_WithDACEnabled_AccessByNonNodeOwner_PublicDocument_CanAccess(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			// Temporarily disable to allow easy creation of public document, and test the disabled nac case.
			testUtils.DisableNAC{
				Identity: testUtils.ClientIdentity(1),
			},
			// Make a non-node-owner own a document.
			testUtils.AddDACPolicy{
				Identity: testUtils.NodeIdentity(0), // Doesn't matter who adds the policy.
				Policy:   examplePolicy,
			},
			&action.AddSchema{
				Identity: testUtils.NoIdentity(),
				Schema:   `type Users @policy(id: "{{.Policy0}}", resource: "users") { name: String }`,
			},
			&action.AddDoc{
				Identity:     testUtils.NoIdentity(),
				CollectionID: 0,
				Doc:          `{ "name": "Shahzad" }`,
			},

			// Public document so anyone can access if NAC is temporarily disabled.
			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
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
