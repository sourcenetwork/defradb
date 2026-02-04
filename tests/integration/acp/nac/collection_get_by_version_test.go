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

	"github.com/sourcenetwork/defradb/client/options"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNAC_GatesCollectionGetByVersion_AuthorizedIdentity_AllowAccess(t *testing.T) {
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
				FilterOptions: options.GetCollections().SetVersionID("does not exist").SetIncludeInactive(false),
				ExpectedError: "key not found", // Note: it is authorized, just key not found.
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesCollectionGetByVersion_NoIdentity_NotAuthorizedError(t *testing.T) {
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
				FilterOptions: options.GetCollections().SetVersionID("does not exist").SetIncludeInactive(false),
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeCollectionGetPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesCollectionGetByVersion_WrongIdentity_NotAuthorizedError(t *testing.T) {
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
				FilterOptions: options.GetCollections().SetVersionID("does not exist").SetIncludeInactive(false),
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeCollectionGetPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
