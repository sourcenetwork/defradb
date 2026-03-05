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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestP2PCollectionGetAll(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.ListP2PCollections{
				NodeID:                0,
				ExpectedCollectionIDs: []int{},
			},
			testUtils.ListP2PCollections{
				NodeID:                1,
				ExpectedCollectionIDs: []int{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
