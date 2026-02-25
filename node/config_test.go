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
	"github.com/sourcenetwork/defradb/internal/utils"
)

func TestSetDisableP2P(t *testing.T) {
	opts := utils.NewOptions(options.Node().SetDisableP2P(true))
	assert.Equal(t, true, opts.DisableP2P)
}

func TestSetDisableAPI(t *testing.T) {
	opts := utils.NewOptions(options.Node().SetDisableAPI(true))
	assert.Equal(t, true, opts.DisableAPI)
}

func TestSetEnableDevelopment(t *testing.T) {
	opts := utils.NewOptions(options.Node().SetEnableDevelopment(true))
	assert.Equal(t, true, opts.EnableDevelopment)
}
