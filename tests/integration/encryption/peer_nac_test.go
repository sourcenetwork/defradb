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

// TestDocEncryptionNAC_SyncBranchableCollection_AuthorizedIdentity_AllowAccess reproduces a bug
// where the KMS pubsub service on the serving peer cannot authorize its own internal collection
// lookup when NAC is enabled and the doc is encrypted: doesIdentityHaveDocPermission calls
// RetrieveCollectionFromDocID with no identity, NAC denies the implicit GetCollections call
// (NodeGetCollectionPerm), the local key fetch errors out, and the requesting peer surfaces
// "failed to retrieve encryption key during DAG sync: EOF".
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
