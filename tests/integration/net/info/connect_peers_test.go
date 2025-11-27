// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package info

import (
	"testing"
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestNetInfoPeers_NoP2PConfigured(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			// Everything besides the JS client is supported, as the JS client does not have
			// an `ActivePeers` function to call.
			[]state.ClientType{
				state.CClientType,
				state.CLIClientType,
				state.GoClientType,
				state.HTTPClientType,
			},
		),
		Actions: []any{
			&action.ActivePeers{
				NodeID:        0,
				ExpectedError: "no p2p system configured",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNetInfoPeers(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			&action.ActivePeers{
				NodeID:   0,
				Expected: []string{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNetInfoConnectPeers(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			&action.ActivePeers{
				NodeID:   0,
				Expected: []string{"{{.Peer1_Address0}}"},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestNetInfoConnectMultiplePeers(t *testing.T) {
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
				SourceNodeID: 2,
				TargetNodeID: 0,
			},
			&action.ActivePeers{
				NodeID: 0,
				Expected: []string{
					"{{.Peer1_Address0}}",
					"{{.Peer2_Address0}}",
				},
			},
			testUtils.Wait{
				// Wait for the connections to propagate
				Duration: time.Millisecond * 50,
			},
			&action.ActivePeers{
				NodeID: 1,
				Expected: []string{
					"{{.Peer0_Address0}}",
					// Node 1 is connected to node 2, because node 0 added them to the same network
					"{{.Peer2_Address0}}",
				},
			},
			&action.ActivePeers{
				NodeID: 2,
				Expected: []string{
					"{{.Peer0_Address0}}",
					// Node 2 is connected to node 1, because node 0 added them to the same network
					"{{.Peer1_Address0}}",
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
