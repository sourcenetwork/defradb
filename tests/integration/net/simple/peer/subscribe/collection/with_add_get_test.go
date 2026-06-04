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

package subscribe_test

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestP2PCollectionAddGetSingle(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
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

func TestP2PCollectionAddGetMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				// Note: If a test is failing here in the error trace, you likely need to change the
				// order of these collection types declared below (some renaming can cause this).
				SDL: `
					type Users {
						name: String
					}
					type Giraffes {
						name: String
					}
					type Bears {
						name: String
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0, 2},
			},
			testUtils.ListP2PCollections{
				NodeID:                1,
				ExpectedCollectionIDs: []int{0, 2},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
