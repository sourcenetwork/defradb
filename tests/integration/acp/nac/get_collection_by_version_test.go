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

	"github.com/sourcenetwork/defradb/client/options"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_GatesGetCollectionByVersion_AuthorizedIdentity_AllowAccess(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// This should work as the identity is authorized.
			&action.GetCollections{
				Identity:      testUtils.ClientIdentity(1),
				FilterOptions: options.GetCollections().SetVersionID("does not exist").SetGetInactive(false),
				ExpectedError: "collection not found", // Note: it is authorized, just collection not found.
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesGetCollectionByVersion_NoIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// We haven't authorized non-identities. So, this should error.
			&action.GetCollections{
				Identity:      testUtils.NoIdentity(),
				FilterOptions: options.GetCollections().SetVersionID("does not exist").SetGetInactive(false),
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetCollectionPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesGetCollectionByVersion_WrongIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// Wrong user/identity will also not be authorized.
			&action.GetCollections{
				Identity:      testUtils.ClientIdentity(2),
				FilterOptions: options.GetCollections().SetVersionID("does not exist").SetGetInactive(false),
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetCollectionPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
