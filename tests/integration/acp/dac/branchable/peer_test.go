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
	"github.com/sourcenetwork/defradb/tests/multiplier"
	"github.com/sourcenetwork/defradb/tests/state"
)

const commitsQuery = `
	query {
		_commits {
			cid
		}
	}
`

// A private branchable collection's commits only sync to a peer node when that node's node identity is
// an actor on the collection object. Here the peer node (node 1) is NOT granted access to the
// collection object, so even a public document's commits (and the collection-level commit) are gated
// at the sync layer and never reach node 1 - the owner sees nothing on the peer.
func TestACP_P2PBranchableCollectionNotSyncedWithoutNodeCollectionAccess_LocalACP(t *testing.T) {
	test := testUtils.TestCase{
		// The document is written without an identity into an ACP gated collection, so
		// the head read-back that works out what the peer should receive is
		// unauthenticated and finds nothing.
		// https://github.com/sourcenetwork/defradb/issues/5196
		MultiplierExcludes: []string{
			multiplier.CrossVersionOldSource,
			multiplier.CrossVersionNewSource,
		},
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

			// Registers the (private) collection object on both nodes, owned by identity 1.
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

			// Public document - but node 1's node identity is not an actor on the collection object,
			// so the sync layer refuses the related blocks.
			&action.AddDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				Doc:          userDoc,
			},

			// Even the collection owner sees nothing on the peer node: the blocks never synced.
			&action.Request{
				NodeID:   immutable.Some(1),
				Identity: testUtils.ClientIdentity(1),
				Request:  commitsQuery,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Once the peer node's node identity is an actor (reader) on the collection object, the private
// branchable collection's commits sync to it. Query-time gating still applies on top: the owner can
// read every commit, but a stranger and an unidentified request see nothing.
func TestACP_P2PBranchableCollectionSyncedWithNodeCollectionAccess_LocalACP(t *testing.T) {
	ownerCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		// The document is written without an identity into an ACP gated collection, so
		// the head read-back that works out what the peer should receive is
		// unauthenticated and finds nothing.
		// https://github.com/sourcenetwork/defradb/issues/5196
		MultiplierExcludes: []string{
			multiplier.CrossVersionOldSource,
			multiplier.CrossVersionNewSource,
		},
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

			// Grant node 1's node identity read access to the collection object so the sync layer will
			// let its related blocks through. Granted on all nodes (Local DAC).
			&action.AddDACCollectionActorRelationship{
				CollectionID:      0,
				Relation:          "reader",
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NodeIdentity(1),
				ExpectedExistence: false,
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

			// Stranger sees nothing on the peer node (query-time gating on the private collection).
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

// With the collection synced to the peer node (via node-identity access), the owner can then share
// read access to the collection commit DAG with a stranger. The relationship is added on every node
// (Local DAC), so the peer node enforces the grant locally and the stranger can read the synced DAG.
func TestACP_P2PBranchableCollectionSharedReaderCanReadOnPeer_LocalACP(t *testing.T) {
	afterCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		// The document is written without an identity into an ACP gated collection, so
		// the head read-back that works out what the peer should receive is
		// unauthenticated and finds nothing.
		// https://github.com/sourcenetwork/defradb/issues/5196
		MultiplierExcludes: []string{
			multiplier.CrossVersionOldSource,
			multiplier.CrossVersionNewSource,
		},
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

			// Let the peer node receive the collection's blocks.
			&action.AddDACCollectionActorRelationship{
				CollectionID:      0,
				Relation:          "reader",
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NodeIdentity(1),
				ExpectedExistence: false,
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
			&action.AddDACCollectionActorRelationship{
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
