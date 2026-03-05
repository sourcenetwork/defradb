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
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestP2PDocument_GetAllWithNoneConfigured_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.ListP2PDocuments{
				NodeID:         0,
				ExpectedDocIDs: []state.ColDocIndex{},
			},
			testUtils.ListP2PDocuments{
				NodeID:         1,
				ExpectedDocIDs: []state.ColDocIndex{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
