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
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/immutable"
)

var (
	userCollectionID  = "bafyreideggss3a43nnydp35fume5nyzlcrqkpo2d7lbvzkdhdeax5gc4cq"
	orderCollectionID = "bafyreihjlbxdbishu6kjbm6ohldh5ypktazuyx66fqy6nldwisud5fthfm"
)

func TestReplicatorList_WithSingleCollectionAndSinglePeer_ShouldSucceed(t *testing.T) {
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
			&action.P2PReplicatorCreate{
				Addresses:   []string{addresses[0]},
				Collections: []string{"User"},
			},
			&action.P2PReplicatorList{
				Expected: immutable.Some([]client.Replicator{
					{
						ID:            peerIDs[0],
						Addresses:     []string{addresses[0]},
						CollectionIDs: []string{userCollectionID},
					},
				}),
			},
		},
	}

	test.Execute(t)
}

func TestReplicatorGetAll_WithMultipleCollectionsAndSinglePeer_ShouldSucceed(t *testing.T) {
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
			&action.P2PReplicatorCreate{
				Addresses:   []string{addresses[0]},
				Collections: []string{"User", "Order"},
			},
			&action.P2PReplicatorList{
				Expected: immutable.Some([]client.Replicator{
					{
						ID:            peerIDs[0],
						Addresses:     []string{addresses[0]},
						CollectionIDs: []string{userCollectionID, orderCollectionID},
					},
				}),
			},
		},
	}

	test.Execute(t)
}

func TestReplicatorGetAll_WithMultipleCollectionsnAndDeleteACollection_ShouldReturnOneCollection(t *testing.T) {
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
			&action.P2PReplicatorCreate{
				Addresses:   []string{addresses[0]},
				Collections: []string{"User", "Order"},
			},
			&action.P2PReplicatorDelete{
				PeerID:      peerIDs[0],
				Collections: []string{"Order"},
			},
			&action.P2PReplicatorList{
				Expected: immutable.Some([]client.Replicator{
					{
						ID:            peerIDs[0],
						Addresses:     []string{addresses[0]},
						CollectionIDs: []string{userCollectionID},
					},
				}),
			},
		},
	}

	test.Execute(t)
}

func TestReplicatorGetAll_WithMultipleCollectionsAndMultiplePeers_ShouldSucceed(t *testing.T) {
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
			&action.P2PReplicatorCreate{
				Addresses:   addresses,
				Collections: []string{"User", "Order"},
			},
			&action.P2PReplicatorList{
				Expected: immutable.Some([]client.Replicator{
					{
						ID:            peerIDs[1],
						Addresses:     []string{addresses[1]},
						CollectionIDs: []string{userCollectionID, orderCollectionID},
					},
					{
						ID:            peerIDs[0],
						Addresses:     []string{addresses[0]},
						CollectionIDs: []string{userCollectionID, orderCollectionID},
					},
				}),
			},
		},
	}

	test.Execute(t)
}

func TestReplicatorGetAll_WithMultiplePeersAndDeleteOfPeer_ShouldReturnOnePeer(t *testing.T) {
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
			&action.P2PReplicatorCreate{
				Addresses:   addresses,
				Collections: []string{"User", "Order"},
			},
			&action.P2PReplicatorDelete{
				PeerID: peerIDs[0],
			},
			&action.P2PReplicatorList{
				Expected: immutable.Some([]client.Replicator{
					{
						ID:            peerIDs[1],
						Addresses:     []string{addresses[1]},
						CollectionIDs: []string{userCollectionID, orderCollectionID},
					},
				}),
			},
		},
	}

	test.Execute(t)
}
