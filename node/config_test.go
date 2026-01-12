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

	"github.com/sourcenetwork/defradb/client/options"
)

func TestApplyNodeOptions_DisableP2P(t *testing.T) {
	config := &Config{}
	nodeOpts := options.Node().SetDisableP2P(true)
	config.applyNodeOptions(nodeOpts)
	assert.Equal(t, true, config.disableP2P)
}

func TestApplyNodeOptions_DisableAPI(t *testing.T) {
	config := &Config{}
	nodeOpts := options.Node().SetDisableAPI(true)
	config.applyNodeOptions(nodeOpts)
	assert.Equal(t, true, config.disableAPI)
}

func TestApplyNodeOptions_EnableDevelopment(t *testing.T) {
	config := &Config{}
	nodeOpts := options.Node().SetEnableDevelopment(true)
	config.applyNodeOptions(nodeOpts)
	assert.Equal(t, true, config.enableDevelopment)
}
