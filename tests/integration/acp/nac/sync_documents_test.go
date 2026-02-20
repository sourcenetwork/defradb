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
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestNAC_GatesSyncDocuments_AuthorizedIdentity_AllowAccess(t *testing.T) {
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

			&action.CreateDoc{
				Identity: testUtils.ClientIdentity(1),
				NodeID:   immutable.Some(0),
				Doc: `{
					"name": "John"
				}`,
			},

			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
				Identity:     testUtils.ClientIdentity(1),
			},

			// This should work as the identity is authorized.
			testUtils.SyncDocs{
				Identity:     testUtils.ClientIdentity(1),
				NodeID:       1,
				CollectionID: 0,
				DocIDs:       []int{0},
				SourceNodes:  []int{0},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesSyncDocuments_NoIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
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

			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type Users {
						name: String
					}
				`,
			},

			&action.CreateDoc{
				Identity: testUtils.ClientIdentity(1),
				Doc: `{
					"name": "John"
				}`,
			},

			// We haven't authorized non-identities. So, this should error.
			testUtils.SyncDocs{
				Identity:      testUtils.NoIdentity(),
				NodeID:        1,
				CollectionID:  0,
				DocIDs:        []int{0},
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeP2PSyncDocumentsPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNAC_GatesSyncDocuments_WrongIdentity_NotAuthorizedError(t *testing.T) {
	test := testUtils.TestCase{
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

			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type Users {
						name: String
					}
				`,
			},

			&action.CreateDoc{
				Identity: testUtils.ClientIdentity(1),
				Doc: `{
					"name": "John"
				}`,
			},

			// Wrong user/identity will also not be authorized.
			testUtils.SyncDocs{
				Identity:      testUtils.ClientIdentity(2),
				NodeID:        1,
				CollectionID:  0,
				DocIDs:        []int{0},
				ExpectedError: testUtils.FormatExpectedErrorWithPermission(acpTypes.NodeP2PSyncDocumentsPerm),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
