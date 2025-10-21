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

func TestReplicatorSet_WithNonExistentCollection_ShouldFail(t *testing.T) {
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
			&action.P2PReplicatorSet{
				Addresses:   []string{addresses[0]},
				Collections: []string{"Order"}, // Non-existent collection
				ExpectError: "failed to get collections for replicator",
			},
		},
	}

	test.Execute(t)
}

func TestReplicatorSet_WithInvalidPeerID_ShouldFail(t *testing.T) {
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
			&action.P2PReplicatorSet{
				Addresses:   []string{addressWithInvalidPeerID},
				Collections: []string{"User"},
				ExpectError: "invalid value \"invalid-peer-id\" for protocol p2",
			},
		},
	}

	test.Execute(t)
}

func TestReplicatorSet_WithInvalidIP_ShouldFail(t *testing.T) {
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
			&action.P2PReplicatorSet{
				Addresses:   []string{addressWithInvalidIP},
				Collections: []string{"User"},
				ExpectError: "invalid value \"999.999.999.999\" for protocol ip4",
			},
		},
	}

	test.Execute(t)
}

func TestReplicatorSet_WithSingleCollectionAndSinglePeer_ShouldSucceed(t *testing.T) {
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
			&action.P2PReplicatorSet{
				Addresses:   []string{addresses[0]},
				Collections: []string{"User"},
			},
		},
	}

	test.Execute(t)
}

func TestReplicatorSet_WithMultipleCollectionsAndSinglePeer_ShouldSucceed(t *testing.T) {
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
			&action.P2PReplicatorSet{
				Addresses:   []string{addresses[0]},
				Collections: []string{"User", "Order"},
			},
		},
	}

	test.Execute(t)
}

func TestReplicatorSet_WithMultipleCollectionsAndMultiplePeers_ShouldSucceed(t *testing.T) {
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
			&action.P2PReplicatorSet{
				Addresses:   addresses,
				Collections: []string{"User", "Order"},
			},
		},
	}

	test.Execute(t)
}
