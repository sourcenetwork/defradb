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

func TestDocEncryptionACP_BranchableCollection_AuthorizedPeerCanFetch(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		// Like the other multi-node branchable tests, this is restricted to local document acp: the
		// collection object is registered per node, whereas a shared SourceHub chain would reject the
		// second node re-registering the same collection object.
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
				Policy:   policy,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users @branchable @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},
			// Let the peer node's node identity receive the collection's gated blocks.
			&action.AddDACCollectionActorRelationship{
				CollectionID:      0,
				Relation:          "reader",
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NodeIdentity(1),
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			// Public (but encrypted) doc — its key is fetchable by anyone who can receive its blocks,
			// which here is gated by the private branchable collection object.
			&action.AddDoc{
				NodeID:         immutable.Some(0),
				CollectionID:   0,
				Doc:            `{"name": "Fred", "age": 33}`,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSync{},
			// The collection owner has access to the collection object, so it reads the decrypted
			// document on the peer node.
			&action.Request{
				NodeID:   immutable.Some(1),
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Fred"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionACP_BranchableCollectionWithSourceHub_AuthorizedPeerCanFetch(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		SupportedDocumentACPTypes: immutable.Some(
			[]state.DocumentACPType{
				state.SourceHubDocumentACPType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   policy,
			},
			&action.AddCollection{
				Identity: testUtils.ClientIdentity(1),
				SDL: `
					type Users @branchable @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},
			// Let the peer node's node identity receive the collection's gated blocks.
			&action.AddDACCollectionActorRelationship{
				NodeID:            immutable.Some(0),
				CollectionID:      0,
				Relation:          "reader",
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NodeIdentity(1),
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			// Public (but encrypted) doc — its key is fetchable by anyone who can receive its blocks,
			// which here is gated by the private branchable collection object.
			&action.AddDoc{
				NodeID:         immutable.Some(0),
				CollectionID:   0,
				Doc:            `{"name": "Fred", "age": 33}`,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSync{},
			// The collection owner has access to the collection object, so it reads the decrypted
			// document on the peer node.
			&action.Request{
				NodeID:   immutable.Some(1),
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Fred"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
