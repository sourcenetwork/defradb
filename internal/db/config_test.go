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

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/immutable"
)

func TestDefaultDBConfig(t *testing.T) {
	cfg := defaultDBConfig()
	assert.True(t, cfg.MaxTxnRetries.HasValue())
	assert.Equal(t, defaultMaxTxnRetries, cfg.MaxTxnRetries.Value())
	assert.True(t, cfg.EnableSigning)
	assert.False(t, cfg.DocumentACP.HasValue())
	assert.False(t, cfg.P2P.HasValue())
}

func TestDBConfigWithMaxRetries(t *testing.T) {
	cfg := defaultDBConfig()
	cfg.MaxTxnRetries = immutable.Some(10)
	assert.True(t, cfg.MaxTxnRetries.HasValue())
	assert.Equal(t, 10, cfg.MaxTxnRetries.Value())
}
