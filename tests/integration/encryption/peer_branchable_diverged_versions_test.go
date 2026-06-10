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
)

// TestDocEncryption_BranchableSync_DivergedVersions_FailsKMSAuth reproduces #4789:
// when each peer has its own diverged @branchable collection version and an
// encrypted doc was created against the originating peer's local version,
// SyncBranchableCollection fails because the receiving peer can't authorize
// its own KMS reply — the doc's CollectionVersionID is part of the DAG bytes
// still in flight.
//
// Symptom: "failed to retrieve encryption key during DAG sync: no peer
// supplied the encryption key".
func TestDocEncryption_BranchableSync_DivergedVersions_FailsKMSAuth(t *testing.T) {
	t.Skip("pending https://github.com/sourcenetwork/defradb/issues/4789")
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User @branchable {
						name: String
					}
				`,
			},
			// Diverge: each peer patches its local @branchable collection
			// independently, producing two different CollectionVersionIDs
			// neither side knows about.
			&action.PatchCollection{
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/User/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			&action.PatchCollection{
				NodeID: immutable.Some(1),
				Patch: `
					[
						{ "op": "add", "path": "/User/Fields/-", "value": {"Name": "score", "Kind": 4} }
					]
				`,
			},
			// Encrypted doc on each peer, against each peer's diverged version.
			&action.AddDoc{
				NodeID:         immutable.Some(0),
				DocMap:         map[string]any{"name": "Andy", "email": "andy@gmail.com"},
				IsDocEncrypted: true,
			},
			&action.AddDoc{
				NodeID:         immutable.Some(1),
				DocMap:         map[string]any{"name": "Fred", "score": 100},
				IsDocEncrypted: true,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			// Both peers sync each other's history (matches the failing test's shape).
			&action.SyncBranchableCollection{
				NodeID:       1,
				CollectionID: 0,
			},
			&action.SyncBranchableCollection{
				NodeID:       0,
				CollectionID: 0,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
