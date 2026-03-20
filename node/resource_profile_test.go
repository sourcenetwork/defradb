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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceLimitsForProfile_Limited(t *testing.T) {
	limits, err := resourceLimitsForProfile(ResourceProfileLimited)
	require.NoError(t, err)
	assert.Equal(t, int64(128<<20), limits.MaxMemory)
}

func TestResourceLimitsForProfile_Server(t *testing.T) {
	limits, err := resourceLimitsForProfile(ResourceProfileServer)
	require.NoError(t, err)
	assert.Equal(t, int64(16<<30), limits.MaxMemory)
}

func TestResourceLimitsForProfile_Unknown(t *testing.T) {
	_, err := resourceLimitsForProfile("unknown")
	assert.Error(t, err)
}
