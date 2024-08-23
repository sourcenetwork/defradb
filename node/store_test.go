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
)

func TestWithStore(t *testing.T) {
	options := &StoreOptions{}
	WithStoreType(MemoryStore)(options)
	assert.Equal(t, MemoryStore, options.store)
}

func TestWithStorePath(t *testing.T) {
	options := &StoreOptions{}
	WithStorePath("test")(options)
	assert.Equal(t, "test", options.path)
}
