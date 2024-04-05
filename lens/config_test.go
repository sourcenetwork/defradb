// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package lens

import (
	"testing"

	"github.com/lens-vm/lens/host-go/runtimes/wasmtime"
	"github.com/stretchr/testify/assert"
)

func TestWithPoolSize(t *testing.T) {
	r := &lensRegistry{}
	WithPoolSize(10)(r)
	assert.Equal(t, 10, r.poolSize)
}

func TestWithRuntime(t *testing.T) {
	r := &lensRegistry{}
	WithRuntime(wasmtime.New())(r)
	assert.NotNil(t, r.runtime)
}
