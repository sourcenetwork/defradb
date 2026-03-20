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
	p2p "github.com/sourcenetwork/go-p2p"
)

const (
	// ResourceProfileLimited applies conservative resource limits suitable for
	// constrained or low-power hardware.
	ResourceProfileLimited = "limited"
	// ResourceProfileServer applies generous resource limits suitable for
	// always-on server nodes.
	ResourceProfileServer = "server"
)

// resourceProfiles maps profile names to their resource limits.
// Connection limits are derived automatically from MaxMemory.
var resourceProfiles = map[string]p2p.ResourceLimits{
	ResourceProfileLimited: {
		MaxMemory:          128 << 20, // 128 MiB
		MaxFileDescriptors: 256,
	},
	ResourceProfileServer: {
		MaxMemory: 8 << 30, // 8 GiB
	},
}

// resourceLimitsForProfile returns the resource limits for the given profile name.
func resourceLimitsForProfile(profile string) (p2p.ResourceLimits, error) {
	limits, ok := resourceProfiles[profile]
	if !ok {
		return p2p.ResourceLimits{}, NewErrUnknownResourceProfile(profile)
	}
	return limits, nil
}
