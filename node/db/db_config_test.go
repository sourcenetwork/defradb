// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"testing"

	"github.com/lens-vm/lens/host-go/runtimes/wasmtime"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestWithACP(t *testing.T) {
	d := &db.Options{}
	WithACP("test")(d)
	assert.True(t, d.ACP.HasValue())
}

func TestWithACPInMemory(t *testing.T) {
	d := &db.Options{}
	WithACPInMemory()(d)
	assert.True(t, d.ACP.HasValue())
}

func TestWithUpdateEvents(t *testing.T) {
	d := &db.Options{}
	WithUpdateEvents()(d)
	assert.NotNil(t, d.Events)
}

func TestWithMaxRetries(t *testing.T) {
	d := &db.Options{}
	WithMaxRetries(10)(d)
	assert.True(t, d.MaxTxnRetries.HasValue())
	assert.Equal(t, 10, d.MaxTxnRetries.Value())
}

func TestWithLensPoolSize(t *testing.T) {
	d := &db.Options{}
	WithLensPoolSize(10)(d)
	assert.Equal(t, 10, d.LensPoolSize.Value())
}

func TestWithLensRuntime(t *testing.T) {
	d := &db.Options{}
	WithLensRuntime(wasmtime.New())(d)
	assert.NotNil(t, d.LensRuntime.Value())
}
