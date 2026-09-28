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

//go:build !js

package state

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/tests/clients"
)

// hostClient stands in for a client wrapper that fronts an HTTP host.
type hostClient struct {
	clients.Client
	host string
}

func (c hostClient) Host() string { return c.host }

func TestGetNodeAudience(t *testing.T) {
	nodes := []*NodeState{
		{Client: hostClient{host: "http://127.0.0.1:1111"}},
		{Client: hostClient{host: "http://127.0.0.1:2222"}},
	}

	tests := []struct {
		name        string
		nodes       []*NodeState
		setupHost   string
		setupNodeID int
		nodeIndex   int
		want        string
		wantSome    bool
	}{
		{
			name:      "reads the host off the node",
			nodes:     nodes,
			nodeIndex: 1,
			want:      "127.0.0.1:2222",
			wantSome:  true,
		},
		{
			// A restarted node gets a new port, and the node rejects tokens minted
			// from the old one.
			name:        "setup host wins for the node being set up",
			nodes:       nodes,
			setupHost:   "http://127.0.0.1:9999",
			setupNodeID: 1,
			nodeIndex:   1,
			want:        "127.0.0.1:9999",
			wantSome:    true,
		},
		{
			name:        "setup host does not leak to other nodes",
			nodes:       nodes,
			setupHost:   "http://127.0.0.1:9999",
			setupNodeID: 1,
			nodeIndex:   0,
			want:        "127.0.0.1:1111",
			wantSome:    true,
		},
		{
			// A node is first set up before it reaches Nodes.
			name:        "setup host answers for a node not on Nodes yet",
			nodes:       nil,
			setupHost:   "http://127.0.0.1:9999",
			setupNodeID: 0,
			nodeIndex:   0,
			want:        "127.0.0.1:9999",
			wantSome:    true,
		},
		{
			name:      "no host and no node gives nothing",
			nodes:     nil,
			nodeIndex: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &State{
				Nodes:              test.nodes,
				CurrentSetupHost:   test.setupHost,
				CurrentSetupNodeID: test.setupNodeID,
			}

			got := GetNodeAudience(s, test.nodeIndex)

			assert.Equal(t, test.wantSome, got.HasValue())
			if test.wantSome {
				assert.Equal(t, test.want, got.Value())
			}
		})
	}
}
