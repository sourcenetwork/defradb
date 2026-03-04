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

	"github.com/sourcenetwork/immutable"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestNAC_GatesGetActivePeers_AuthorizedIdentity_AllowAccess(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.HTTPClientType,
				state.CLIClientType,
				state.GoClientType,
				state.CClientType,
			},
		),
		Actions: []any{
			// Doing this in the beggining is important to start all nodes with NAC enabled.
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			// Starting all nodes with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// This should work as the identity is authorized.
			&action.ActivePeers{
				Identity: testUtils.ClientIdentity(1),
				NodeID:   1,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesGetActivePeers_NoIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Doing this in the beggining is important to start all nodes with NAC enabled.
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			// Starting all nodes with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// We haven't authorized non-identities. So, this should error.
			&action.ActivePeers{
				Identity:      testUtils.NoIdentity(),
				NodeID:        1,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetP2PActivePeersPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesGetActivePeers_WrongIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			// Doing this in the beggining is important to start all nodes with NAC enabled.
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			// Starting all nodes with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// Wrong user/identity will also not be authorized.
			&action.ActivePeers{
				Identity:      testUtils.ClientIdentity(2),
				NodeID:        1,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeGetP2PActivePeersPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
