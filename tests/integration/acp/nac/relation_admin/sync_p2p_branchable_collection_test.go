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

package test_acp_nac_relation_admin

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestNAC_AdminRelation_CanSyncP2PBranchableCollection(t *testing.T) {
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
			// Doing this in the beginning is important to start all nodes with NAC enabled.
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			// Starting all nodes with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			testUtils.ConnectPeers{
				Identity:     testUtils.ClientIdentity(1),
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type User @branchable {
						name: String
					}
				`,
			},
			&action.AddDoc{
				NodeID:   immutable.Some(0),
				Identity: testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name": "John",
				},
			},

			// This user, can not perform this gated operation yet.
			&action.SyncBranchableCollection{
				Identity:      testUtils.ClientIdentity(2),
				NodeID:        1,
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeSyncP2PBranchableCollectionPerm),
			},

			// Grant access to user.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// This user, can now perform this gated operation.
			&action.SyncBranchableCollection{
				Identity: testUtils.ClientIdentity(2),
				NodeID:   1,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
