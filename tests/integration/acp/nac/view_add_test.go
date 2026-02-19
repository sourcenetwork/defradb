// Copyright 2026 Democratized Data Foundation
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

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_GatesViewAdd_AuthorizedIdentity_AllowAccess(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type Users {
						name: String
					}
				`,
			},

			// This should work as the identity is authorized.
			&action.CreateView{
				Identity: testUtils.ClientIdentity(1),
				Query: `
					Users {
						name
					}
				`,
				SDL: `
					type UsersView @materialized(if: false) {
						name: String
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesViewAdd_NoIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type Users {
						name: String
					}
				`,
			},

			// We haven't authorized non-identities. So, this should error.
			&action.CreateView{
				Identity: testUtils.NoIdentity(),
				Query: `
					Users {
						name
					}
				`,
				SDL: `
					type UsersView @materialized(if: false) {
						name: String
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeViewAddPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesViewAdd_WrongIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type Users {
						name: String
					}
				`,
			},

			// Wrong user/identity will also not be authorized.
			&action.CreateView{
				Identity: testUtils.ClientIdentity(2),
				Query: `
					Users {
						name
					}
				`,
				SDL: `
					type UsersView @materialized(if: false) {
						name: String
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeViewAddPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
