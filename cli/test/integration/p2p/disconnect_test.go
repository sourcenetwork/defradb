// Copyright 2026 Democratized Data Foundation
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
	"testing"

	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
)

func TestDisconnect_WithInvalidPeerID_ShouldFail(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.DisconnectP2P{
				Addresses:   []string{addressWithInvalidPeerID},
				ExpectError: "invalid value \"invalid-peer-id\" for protocol p2",
			},
		},
	}

	test.Execute(t)
}

func TestDisconnect_WithInvalidIP_ShouldFail(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.DisconnectP2P{
				Addresses:   []string{addressWithInvalidIP},
				ExpectError: "invalid value \"999.999.999.999\" for protocol ip4",
			},
		},
	}

	test.Execute(t)
}

func TestDisconnect_WithSinglePeer_ShouldSucceed(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.DisconnectP2P{
				Addresses: []string{addresses[0]},
			},
		},
	}

	test.Execute(t)
}

func TestDisconnect_WithMultiplePeers_ShouldSucceed(t *testing.T) {
	test := &integration.Test{
		Actions: []action.Action{
			&action.DisconnectP2P{
				Addresses: addresses,
			},
		},
	}

	test.Execute(t)
}
