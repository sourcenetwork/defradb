// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"testing"

	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
)

func TestReplicatorDelete_WithNonExistentCollection_ShouldFail(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.SchemaAdd{
				InlineSchema: `
					type User {
						name: String
						age: Int
						email: String
					}
				`,
			},
			&action.P2PReplicatorAdd{
				Addresses:   []string{addresses[0]},
				Collections: []string{"User"},
			},
			&action.P2PReplicatorDelete{
				PeerID:      peerIDs[0],
				Collections: []string{"Order"}, // Non-existent collection
				ExpectError: "failed to get collections for replicator",
			},
		},
	}

	test.Execute(t)
}

func TestReplicatorDelete_WithInvalidPeerID_ShouldFail(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.SchemaAdd{
				InlineSchema: `
					type User {
						name: String
						age: Int
						email: String
					}
				`,
			},
			&action.P2PReplicatorAdd{
				Addresses:   []string{addresses[0]},
				Collections: []string{"User"},
			},
			&action.P2PReplicatorDelete{
				PeerID:      invalidPeerID,
				Collections: []string{"User"},
				ExpectError: "replicator not found",
			},
		},
	}

	test.Execute(t)
}

func TestReplicatorDelete_WithSingleCollectionAndSinglePeer_ShouldSucceed(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.SchemaAdd{
				InlineSchema: `
					type User {
						name: String
						age: Int
						email: String
					}
				`,
			},
			&action.P2PReplicatorAdd{
				Addresses:   []string{addresses[0]},
				Collections: []string{"User"},
			},
			&action.P2PReplicatorDelete{
				PeerID:      peerIDs[0],
				Collections: []string{"User"},
			},
		},
	}

	test.Execute(t)
}

func TestReplicatorDelete_WithMultiplePeersDeleteSinglePeer_ShouldSucceed(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.SchemaAdd{
				InlineSchema: `
					type User {
						name: String
						age: Int
						email: String
					}
					type Order {
						orderNumber: String
						amount: Float
					}
				`,
			},
			&action.P2PReplicatorAdd{
				Addresses:   addresses,
				Collections: []string{"User", "Order"},
			},
			&action.P2PReplicatorDelete{
				PeerID: peerIDs[0],
			},
		},
	}

	test.Execute(t)
}

func TestReplicatorDelete_WithMultipleCollectionsDeleteSingeCollection_ShouldSucceed(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.SchemaAdd{
				InlineSchema: `
					type User {
						name: String
						age: Int
						email: String
					}
					type Order {
						orderNumber: String
						amount: Float
					}
				`,
			},
			&action.P2PReplicatorAdd{
				Addresses:   addresses,
				Collections: []string{"User", "Order"},
			},
			&action.P2PReplicatorDelete{
				PeerID:      peerIDs[0],
				Collections: []string{"User"},
			},
		},
	}

	test.Execute(t)
}
