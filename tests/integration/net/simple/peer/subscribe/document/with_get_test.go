// Copyright 2025 Democratized Data Foundation
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

func TestP2PDocument_GetAllWithNoneConfigured_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.GetAllP2PDocuments{
				NodeID:         0,
				ExpectedDocIDs: []testUtils.ColDocIndex{},
			},
			testUtils.GetAllP2PDocuments{
				NodeID:         1,
				ExpectedDocIDs: []testUtils.ColDocIndex{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
