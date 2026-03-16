// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// P2P networking stack does not work in JS builds.
//
//go:build !js

package node

import (
	"fmt"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
)

const (
	// ResourceProfileLimited applies conservative resource limits suitable for constrained hardware such as edge nodes
	ResourceProfileLimited = "limited"
	// ResourceProfileServer only applies limits per-peer, rest is autoscaled
	ResourceProfileServer = "server"
)

// resourceProfiles maps profile names to their PartialLimitConfig overrides.
var resourceProfiles = map[string]rcmgr.PartialLimitConfig{
	ResourceProfileLimited: {
		System: rcmgr.ResourceLimits{
			ConnsInbound:    32,
			ConnsOutbound:   64,
			Conns:           96,
			StreamsInbound:  32 * 16,
			StreamsOutbound: 64 * 16,
			Streams:         96 * 16,
			Memory:          128 << 20, // minimal requirement
			FD:              256,       // minimal requirement
		},
		Transient: rcmgr.ResourceLimits{
			ConnsInbound:    16,
			ConnsOutbound:   32,
			Conns:           48,
			StreamsInbound:  64,
			StreamsOutbound: 128,
			Streams:         196,
			Memory:          16 << 20,
			FD:              32,
		},
		PeerDefault: rcmgr.ResourceLimits{
			ConnsInbound:    4,
			ConnsOutbound:   4,
			Conns:           8,
			StreamsInbound:  64,
			StreamsOutbound: 128,
			Streams:         196,
			Memory:          64 << 20,
		},
	},
	ResourceProfileServer: {
		PeerDefault: rcmgr.ResourceLimits{
			ConnsInbound:    8,
			ConnsOutbound:   8,
			Conns:           8,
			StreamsInbound:  512,
			StreamsOutbound: 1024,
			Streams:         1024,
			Memory:          128 << 20,
		},
	},
}

// buildResourceManager constructs a resource manager from the given profile name.
// If profile is empty, nil is returned and go-p2p will use libp2p's autoscaled defaults.
func buildResourceManager(profile string) (network.ResourceManager, error) {
	partial, ok := resourceProfiles[profile]
	if !ok {
		return nil, fmt.Errorf("unknown resource profile %q: valid values are %q, %q",
			profile, ResourceProfileLimited, ResourceProfileServer)
	}
	limits := rcmgr.DefaultLimits
	libp2p.SetDefaultServiceLimits(&limits)
	rm, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(partial.Build(limits.AutoScale())))
	if err != nil {
		return nil, err
	}
	return rm, nil
}
