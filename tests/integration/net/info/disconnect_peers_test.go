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

package info

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestNetInfoDisconnectSinglePeer(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			&action.ActivePeers{
				NodeID:   1,
				Expected: []string{"{{.Peer0_Address0}}"},
			},
			testUtils.DisconnectPeers{
				SourceNodeID:  1,
				TargetNodeIDs: []int{0},
			},
			&action.ActivePeers{
				NodeID:   1,
				Expected: []string{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNetInfoDisconnectMultiplePeers(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 2,
			},
			&action.ActivePeers{
				NodeID:   1,
				Expected: []string{"{{.Peer0_Address0}}", "{{.Peer2_Address0}}"},
			},
			testUtils.DisconnectPeers{
				SourceNodeID:  1,
				TargetNodeIDs: []int{0, 2},
			},
			&action.ActivePeers{
				NodeID:   1,
				Expected: []string{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNetInfoDisconnectEmptyList(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.DisconnectPeers{
				SourceNodeID:  0,
				TargetNodeIDs: []int{},
				ExpectedError: "addresses cannot be empty",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
