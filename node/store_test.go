// Copyright 2024 Democratized Data Foundation
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

func TestSetStoreType(t *testing.T) {
	opts := utils.NewOptions(options.Node().Store().SetType(options.NodeMemoryStore).Node())
	assert.Equal(t, options.NodeMemoryStore, opts.Store.Store)
}

func TestSetStorePath(t *testing.T) {
	opts := utils.NewOptions(options.Node().Store().SetPath("test").Node())
	assert.Equal(t, "test", opts.Store.Path)
}
