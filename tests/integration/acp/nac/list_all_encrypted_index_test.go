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

	"github.com/sourcenetwork/immutable"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestNAC_GatesListAllEncryptedIndex_AuthorizedIdentity_AllowAccess(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.GoClientType,
				state.HTTPClientType,
				state.CLIClientType,
				state.CClientType,
			},
		),
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.ListAllEncryptedIndexes{
				Identity:        testUtils.ClientIdentity(1),
				ExpectedIndexes: map[client.CollectionName][]client.EncryptedIndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesListAllEncryptedIndex_NoIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.ListAllEncryptedIndexes{
				Identity:      testUtils.NoIdentity(),
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeListAllEncryptedIndexPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesListAllEncryptedIndex_WrongIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			testUtils.ListAllEncryptedIndexes{
				Identity:      testUtils.ClientIdentity(2),
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeListAllEncryptedIndexPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
