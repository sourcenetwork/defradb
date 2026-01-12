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
)

func TestNodeStoreOptions_Store(t *testing.T) {
	storeOpts := options.NodeStore()
	storeOpts.Store = options.NodeMemoryStore
	assert.Equal(t, options.NodeMemoryStore, storeOpts.Store)
}

func TestNodeStoreOptions_Path(t *testing.T) {
	storeOpts := options.NodeStore()
	storeOpts.Path = "test"
	assert.Equal(t, "test", storeOpts.Path)
}
