// Copyright 2025 Democratized Data Foundation
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

	"github.com/stretchr/testify/assert"
)

func TestWithDisableP2P(t *testing.T) {
	options := &Config{}
	WithDisableP2P(true)(options)
	assert.Equal(t, true, options.disableP2P)
}

func TestWithDisableAPI(t *testing.T) {
	options := &Config{}
	WithDisableAPI(true)(options)
	assert.Equal(t, true, options.disableAPI)
}

func TestWithEnableDevelopment(t *testing.T) {
	options := &Config{}
	WithEnableDevelopment(true)(options)
	assert.Equal(t, true, options.enableDevelopment)
}
