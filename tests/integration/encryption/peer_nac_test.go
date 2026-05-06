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

package encryption

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// TestDocEncryptionNAC_SyncBranchableCollection_AuthorizedIdentity_AllowAccess
// covers branchable-collection sync of an encrypted doc when NAC is enabled
// on the serving node and the requester carries an authorized identity: the
// requester must receive the key and complete the sync.
//
// Edge case: the serving node's KMS performs an internal collection lookup
// while answering the key request, which crosses NAC's collection-access
// check. That lookup must be authorized as a node-internal call rather than
// be subjected to the requester's NAC permissions.
func TestDocEncryptionNAC_SyncBranchableCollection_AuthorizedIdentity_AllowAccess(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.GoClientType,
				state.HTTPClientType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},
			testUtils.ConnectPeers{
				Identity:     testUtils.ClientIdentity(1),
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
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
				IsDocEncrypted: true,
			},

			&action.SyncBranchableCollection{
				Identity: testUtils.ClientIdentity(1),
				NodeID:   1,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
