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

package test_acp_nac

import (
	"testing"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_GatesAddView_AuthorizedIdentity_AllowAccess(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {
						name: String
					}
				`,
			},

			// This should work as the identity is authorized.
			&action.AddView{
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

func TestNAC_GatesAddView_NoIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {
						name: String
					}
				`,
			},

			// We haven't authorized non-identities. So, this should error.
			&action.AddView{
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
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeAddViewPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesAddView_WrongIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {
						name: String
					}
				`,
			},

			// Wrong user/identity will also not be authorized.
			&action.AddView{
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
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeAddViewPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
