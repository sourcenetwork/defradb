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

func TestNAC_GatesDocumentUpdate_AuthorizedIdentity_AllowAccess(t *testing.T) {
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
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type User {
						name: String
						age: Int 
					}`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc: `{
					"name": "Shahzad",
					"age": 48
				}`,
			},

			// This should work as the identity is authorized.
			testUtils.UpdateDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				DocID:        0,
				Doc:          `{ "name": "Lone" }`,
			},
			&action.Request{ // Should now be updated
				Identity: testUtils.ClientIdentity(1),
				Request:  `query{ User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Lone"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesDocumentUpdate_NoIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		// todo: Investigate and test this behavior across all client types when implementing granular NAC permissions.
		// See: https://github.com/sourcenetwork/defradb/issues/4383
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type User {
						name: String
						age: Int 
					}`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc: `{
					"name": "Shahzad",
					"age": 48
				}`,
			},

			// We haven't authorized non-identities. So, this should error.
			testUtils.UpdateDoc{
				Identity:     testUtils.NoIdentity(),
				CollectionID: 0,
				DocID:        0,
				Doc:          `{ "name": "Lone" }`,
				// todo: After implementing granular NAC permissions, this should be changed to a
				// specific permission error. Currently, the permission error is different across
				// different client types and environments.
				// See: https://github.com/sourcenetwork/defradb/issues/4446
				ExpectedError: "not authorized to perform operation",
			},
			&action.Request{ // Should not be updated
				Identity: testUtils.ClientIdentity(1),
				Request:  `query{ User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Shahzad"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesDocumentUpdate_WrongIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		// todo: Investigate and test this behavior across all client types when implementing granular NAC permissions.
		// See: https://github.com/sourcenetwork/defradb/issues/4383
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type User {
						name: String
						age: Int 
					}`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Doc: `{
					"name": "Shahzad",
					"age": 48
				}`,
			},

			// Wrong user/identity will also not be authorized.
			testUtils.UpdateDoc{
				Identity:     testUtils.ClientIdentity(2),
				CollectionID: 0,
				DocID:        0,
				Doc:          `{ "name": "Lone" }`,
				// todo: After implementing granular NAC permissions, this should be changed to a
				// specific permission error. Currently, the permission error is different across
				// different client types and environments.
				// See: https://github.com/sourcenetwork/defradb/issues/4446
				ExpectedError: "not authorized to perform operation",
			},
			&action.Request{ // Should not be updated
				Identity: testUtils.ClientIdentity(1),
				Request:  `query{ User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Shahzad"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
