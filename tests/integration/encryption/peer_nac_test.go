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

// Branchable collection, encrypted doc, NAC enabled, requester has an
// authorized identity. Sync must succeed end-to-end.
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

// Branchable collection, encrypted doc, NAC enabled, requester carries an
// unauthorized identity. Sync must be denied.
func TestDocEncryptionNAC_SyncBranchableCollection_UnauthorizedIdentity_DenyAccess(t *testing.T) {
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
				Identity:      testUtils.ClientIdentity(2),
				NodeID:        1,
				ExpectedError: "not authorized to perform operation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Branchable collection, encrypted doc, NAC enabled, request carries no
// identity. Sync must be denied.
func TestDocEncryptionNAC_SyncBranchableCollection_NoIdentity_DenyAccess(t *testing.T) {
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
				Identity:      testUtils.NoIdentity(),
				NodeID:        1,
				ExpectedError: "not authorized to perform operation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Branchable collection, encrypted doc, NAC enabled. A non-admin identity is
// initially denied, then granted an admin relation, after which the sync
// succeeds end-to-end.
func TestDocEncryptionNAC_SyncBranchableCollection_GrantedRelation_AllowAccess(t *testing.T) {
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

			// Without a NAC grant, ClientIdentity(2) is denied.
			&action.SyncBranchableCollection{
				Identity:      testUtils.ClientIdentity(2),
				NodeID:        1,
				ExpectedError: "not authorized to perform operation",
			},

			// Grant admin to ClientIdentity(2) on both nodes.
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedExistence: false,
			},

			// Now the same identity can sync end-to-end.
			&action.SyncBranchableCollection{
				Identity: testUtils.ClientIdentity(2),
				NodeID:   1,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Branchable collection, encrypted doc, NAC enabled. A non-admin identity is
// granted an admin relation, syncs successfully, then the relation is
// revoked. The subsequent sync must be denied.
func TestDocEncryptionNAC_SyncBranchableCollection_RevokedRelation_DenyAccess(t *testing.T) {
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

			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				Relation:          "admin",
				ExpectedExistence: false,
			},
			&action.SyncBranchableCollection{
				Identity: testUtils.ClientIdentity(2),
				NodeID:   1,
			},

			testUtils.DeleteNACActorRelationship{
				RequestorIdentity:   testUtils.ClientIdentity(1),
				TargetIdentity:      testUtils.ClientIdentity(2),
				Relation:            "admin",
				ExpectedRecordFound: true,
			},
			&action.SyncBranchableCollection{
				Identity:      testUtils.ClientIdentity(2),
				NodeID:        1,
				ExpectedError: "not authorized to perform operation",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Non-branchable collection, encrypted doc, NAC enabled, gossip-driven sync via
// AddCollectionSubscription. The subscriber's node identity must be granted
// access so the gossip-triggered KMS key fetch (which has no user identity in
// ctx and falls back to the requester's nodeIdentity) passes NAC on the
// publishing peer.
func TestDocEncryptionNAC_GossipSync_AuthorizedNodeIdentity_AllowAccess(t *testing.T) {
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
				SDL:      userCollection,
			},
			testUtils.AddCollectionSubscription{
				Identity:      testUtils.ClientIdentity(1),
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.AddNACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NodeIdentity(1),
				Relation:          "admin",
				ExpectedExistence: false,
			},
			&action.AddDoc{
				NodeID:         immutable.Some(0),
				Identity:       testUtils.ClientIdentity(1),
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID:   immutable.Some(1),
				Identity: testUtils.ClientIdentity(1),
				Request: `query {
					Users {
						age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"age": int64(21)},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
