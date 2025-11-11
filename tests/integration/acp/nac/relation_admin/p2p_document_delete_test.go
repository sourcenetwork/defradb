// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_nac_relation_admin

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

func TestNAC_AdminRelation_CanP2PDocumentDelete(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.HTTPClientType,
				state.CLIClientType,
				state.GoClientType,
				// TODO: https://github.com/sourcenetwork/defradb/issues/4091
				// We have to fix the c-binding identity passing issue to support c-client.
				// state.CClientType,
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
			// Note: Doing setup steps after starting with nac enabled, otherwise the in-memory tests
			// will lose setup state when the restart happens (i.e. the restart that started nac).
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name": "Shahzad Lone",
				},
			},
			testUtils.ConnectPeers{
				Identity:     testUtils.ClientIdentity(1),
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToDocument{
				Identity: testUtils.ClientIdentity(1),
				NodeID:   1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},

			// This user, can not perform this gated operation yet.
			testUtils.UnsubscribeToDocument{
				Identity: testUtils.ClientIdentity(2),
				NodeID:   1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
				ExpectedError: "not authorized to perform operation",
			},

			// Grant access to user.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2), // Grant this user "admin" relation
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// This user, can now perform this gated operation.
			testUtils.UnsubscribeToDocument{
				Identity: testUtils.ClientIdentity(2),
				NodeID:   1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
