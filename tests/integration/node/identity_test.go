// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNodeIdentity_NodeIdentity_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.GetNodeIdentity{
				NodeID:           0,
				ExpectedIdentity: testUtils.NodeIdentity(0),
			},
			testUtils.GetNodeIdentity{
				NodeID:           1,
				ExpectedIdentity: testUtils.NodeIdentity(1),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
