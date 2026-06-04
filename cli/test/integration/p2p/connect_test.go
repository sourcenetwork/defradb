// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
)

var (
	peerIDs = []string{
		"12D3KooWGj2wWqxSKB2Lqg6K6ye7N7W4YcpoeNpdXakHkgGjUqHC",
		"12D3KooWBVR4AGs7pTwdjUH1Pd9bgh85hyf1xtV3rg7bj3VfZGoT",
	}
	invalidPeerID = "invalid-peer-id"
	addresses     = []string{
		fmt.Sprintf("/ip4/127.0.0.1/tcp/9000/p2p/%s", peerIDs[0]),
		fmt.Sprintf("/ip4/127.0.0.1/tcp/9001/p2p/%s", peerIDs[1]),
	}
	addressWithInvalidPeerID = fmt.Sprintf("/ip4/127.0.0.1/tcp/9000/p2p/%s", invalidPeerID)
	addressWithInvalidIP     = "/ip4/999.999.999.999/tcp/9000/p2p/12D3KooWGj2wWqxSKB2Lqg6K6ye7N7W4YcpoeNpdXakHkgGjUqHC"
)

func TestConnect_WithInvalidPeerID_ShouldFail(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.ConnectP2P{
				Addresses:   []string{addressWithInvalidPeerID},
				ExpectError: "invalid value \"invalid-peer-id\" for protocol p2",
			},
		},
	}

	test.Execute(t)
}

func TestConnect_WithInvalidIP_ShouldFail(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.ConnectP2P{
				Addresses:   []string{addressWithInvalidIP},
				ExpectError: "invalid value \"999.999.999.999\" for protocol ip4",
			},
		},
	}

	test.Execute(t)
}

// NOTE: This test currently fails because there is no peer listening at the given address.
// However, it does at least verify that a single address can be passed in.
//
// TODO: Add capability to have multiple defradb instances in tests, so we can
// actually test successful connections. https://github.com/sourcenetwork/defradb/issues/4021
func TestConnect_WithSinglePeer_ShouldSucceed(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.ConnectP2P{
				Addresses:   []string{addresses[0]},
				ExpectError: "connect: connection refused",
			},
		},
	}

	test.Execute(t)
}

// NOTE: This test currently fails because there is no peer listening at the given address.
// However, it does at least verify that multiple addresses can be passed in.
//
// TODO: Add capability to have multiple defradb instances in tests, so we can
// actually test successful connections. https://github.com/sourcenetwork/defradb/issues/4021
func TestConnect_WithMultiplePeers_ShouldSucceed(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.ConnectP2P{
				Addresses:   addresses,
				ExpectError: "connect: connection refused",
			},
		},
	}

	test.Execute(t)
}
