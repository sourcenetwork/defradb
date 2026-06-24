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

package test_acp_dac_branchable

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

const commitsQuery = `
	query {
		_commits {
			cid
		}
	}
`

// A branchable, permissioned collection is registered as an acp object on every node it is added to
// (the collection id is deterministic), so collection-level gating is enforced independently on each
// node even with local document ACP.
//
// Here a public document is created on node 0 and synced to the subscribing node 1. Even though the
// document (and its collection-level commit DAG) replicated to node 1, node 1 still gates the whole
// DAG on its own (private) collection object: the owner can read every commit, but a stranger and an
// unidentified request see nothing.
func TestACP_P2PBranchableCollectionPrivateCollectionGatedOnPeer_LocalACP(t *testing.T) {
	ownerCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		SupportedDocumentACPTypes: immutable.Some(
			[]state.DocumentACPType{
				state.LocalDocumentACPType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),

			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   usersPolicy,
			},

			// Added with identity on all nodes -> the collection object is registered (private) on
			// both node 0 and node 1, owned by identity 1.
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL:      branchablePermissionedSDL,
			},

			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},

			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},

			// Public document - it (and the collection-level commit) freely replicates to node 1.
			&action.AddDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				Doc:          userDoc,
			},

			testUtils.WaitForSync{},

			// Owner reads all 4 commits on the peer node.
			&action.Request{
				NodeID:            immutable.Some(1),
				Identity:          testUtils.ClientIdentity(1),
				Request:           commitsQuery,
				NonOrderedResults: true,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"cid": ownerCid},
						{"cid": ownerCid},
						{"cid": ownerCid},
						{"cid": ownerCid},
					},
				},
			},

			// Stranger sees nothing on the peer node.
			&action.Request{
				NodeID:   immutable.Some(1),
				Identity: testUtils.ClientIdentity(2),
				Request:  commitsQuery,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},

			// Unidentified request sees nothing on the peer node.
			&action.Request{
				NodeID:  immutable.Some(1),
				Request: commitsQuery,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// After the owner grants a stranger the "reader" relation on the collection object, the stranger can
// read the replicated commit DAG on the peer node too. The relationship is added on every node (local
// ACP), so node 1 enforces the grant locally.
func TestACP_P2PBranchableCollectionSharedReaderCanReadOnPeer_LocalACP(t *testing.T) {
	afterCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		SupportedDocumentACPTypes: immutable.Some(
			[]state.DocumentACPType{
				state.LocalDocumentACPType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),

			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   usersPolicy,
			},

			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL:      branchablePermissionedSDL,
			},

			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},

			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},

			&action.AddDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				Doc:          userDoc,
			},

			testUtils.WaitForSync{},

			// Before sharing: stranger sees nothing on the peer node.
			&action.Request{
				NodeID:   immutable.Some(1),
				Identity: testUtils.ClientIdentity(2),
				Request:  commitsQuery,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},

			// Owner shares read access to the collection commit DAG with the stranger (on all nodes).
			testUtils.AddDACCollectionActorRelationship{
				CollectionID:      0,
				Relation:          "reader",
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				ExpectedExistence: false,
			},

			// After sharing: stranger reads all 4 commits on the peer node.
			&action.Request{
				NodeID:            immutable.Some(1),
				Identity:          testUtils.ClientIdentity(2),
				Request:           commitsQuery,
				NonOrderedResults: true,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"cid": afterCid},
						{"cid": afterCid},
						{"cid": afterCid},
						{"cid": afterCid},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
