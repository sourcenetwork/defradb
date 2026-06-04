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
