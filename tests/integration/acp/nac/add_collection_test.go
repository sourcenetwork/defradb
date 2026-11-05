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

func TestNAC_GatesAddCollection_AllowIfAuthorizedElseError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// We haven't authorized non-identities. So, this should error.
			&action.AddCollection{
				Identity: testUtils.NoIdentity(),
				SDL: `
					type Users {
						name: String
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodePatchCollectionPerm),
			},

			// Wrong user/identity will also not be authorized.
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(2),
				SDL: `
					type Users {
						name: String
					}
				`,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodePatchCollectionPerm),
			},

			// This should work as the identity is authorized.
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users {
						name: String
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
