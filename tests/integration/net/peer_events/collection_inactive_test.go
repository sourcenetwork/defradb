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

package peer_events

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestPeerEvents_SubscribeToCollectionWithNoActiveVersion_ShouldSucceed tests that a node can
// subscribe to a collection that has been synced over P2P but has not yet been activated.
func TestPeerEvents_SubscribeToCollectionWithNoActiveVersion_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			// Collection will be active on Node 0
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type User {
						name: String
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			// Collection will arrive inactive on Node 1
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"{{.CollectionVersionID0}}"},
			},
			testUtils.WaitForSync{},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestPeerEvents_DeleteSubscriptionOnCollectionWithNoActiveVersion_ShouldSucceed tests that a
// subscription can be removed on a node where the collection has no active version.
func TestPeerEvents_DeleteSubscriptionOnCollectionWithNoActiveVersion_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			// Collection will be active on Node 0
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type User {
						name: String
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			// Collection will arrive inactive on Node 1
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"{{.CollectionVersionID0}}"},
			},
			testUtils.WaitForSync{},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.DeleteCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestPeerEvents_ListCollectionsWithNoActiveVersion_ShouldSucceed tests that listing P2P
// subscriptions works on a node where the subscribed collection has no active version.
func TestPeerEvents_ListCollectionsWithNoActiveVersion_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			// Collection will be active on Node 0
			&action.AddCollection{
				NodeID: immutable.Some(0),
				SDL: `
					type User {
						name: String
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			// Collection will arrive inactive on Node 1
			&action.SyncCollectionVersions{
				NodeID:     1,
				VersionIDs: []string{"{{.CollectionVersionID0}}"},
			},
			testUtils.WaitForSync{},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.ListP2PCollections{
				NodeID:                1,
				ExpectedCollectionIDs: []int{0},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
