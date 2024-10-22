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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNodeIdentity_NodeIdentity_Succeed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.AssignNodeIdentity{
				NodeID: 0,
				Name:   "node1",
			},
			testUtils.AssignNodeIdentity{
				NodeID: 1,
				Name:   "node2",
			},
			testUtils.GetNodeIdentity{
				NodeID:               0,
				ExpectedIdentityName: immutable.Some("node1"),
			},
			testUtils.GetNodeIdentity{
				NodeID:               1,
				ExpectedIdentityName: immutable.Some("node2"),
			},
			testUtils.AssignNodeIdentity{
				NodeID: 0,
				Name:   "node1_new",
			},
			testUtils.GetNodeIdentity{
				NodeID:               0,
				ExpectedIdentityName: immutable.Some("node1_new"),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
