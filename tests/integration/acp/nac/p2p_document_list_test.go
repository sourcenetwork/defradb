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
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

func TestNAC_GatesP2PDocumentList_AuthorizedIdentity_AllowAccess(t *testing.T) {
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

			// This should work as the identity is authorized.
			testUtils.GetAllP2PDocuments{
				Identity: testUtils.ClientIdentity(1),
				NodeID:   1,
				ExpectedDocIDs: []state.ColDocIndex{
					state.NewColDocIndex(0, 0),
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesP2PDocumentList_NoIdentity_NotAuthorizedError(t *testing.T) {
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

			// We haven't authorized non-identities. So, this should error.
			testUtils.GetAllP2PDocuments{
				Identity:      testUtils.NoIdentity(),
				NodeID:        1,
				ExpectedError: "not authorized to perform operation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesP2PDocumentList_WrongIdentity_NotAuthorizedError(t *testing.T) {
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

			// Wrong user/identity will also not be authorized.
			testUtils.GetAllP2PDocuments{
				Identity:      testUtils.ClientIdentity(2),
				NodeID:        1,
				ExpectedError: "not authorized to perform operation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
