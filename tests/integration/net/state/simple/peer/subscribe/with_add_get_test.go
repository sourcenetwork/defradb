// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package subscribe_test

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestP2PSubscribeAddGetSingle(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.GetAllP2PCollections{
				NodeID:                1,
				ExpectedCollectionIDs: []int{0},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PSubscribeAddGetMultiple(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				// Note: If a test is failing here in the error trace, you likely need to change the
				// order of these schema types declared below (some renaming can cause this).
				Schema: `
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
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0, 2},
			},
			testUtils.GetAllP2PCollections{
				NodeID:                1,
				ExpectedCollectionIDs: []int{2, 0},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
