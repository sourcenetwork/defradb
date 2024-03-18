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
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithInMemory(t *testing.T) {
	options := &StoreOptions{}
	WithInMemory(true)(options)
	assert.Equal(t, true, options.inMemory)
}

func TestWithPath(t *testing.T) {
	options := &StoreOptions{}
	WithPath("tmp")(options)
	assert.Equal(t, "tmp", options.path)
}

func TestWithValueLogFileSize(t *testing.T) {
	options := &StoreOptions{}
	WithValueLogFileSize(int64(5 << 30))(options)
	assert.Equal(t, int64(5<<30), options.valueLogFileSize)
}

func TestWithEncryptionKey(t *testing.T) {
	keyBytes := make([]byte, 32)
	_, err := rand.Read(keyBytes)
	require.NoError(t, err)

	keyHex := hex.EncodeToString(keyBytes)

	options := &StoreOptions{}
	WithEncryptionKey(keyHex)(options)
	assert.Equal(t, keyHex, options.encryptionKey)
}

func TestWithIndexCacheSize(t *testing.T) {
	options := &StoreOptions{}
	WithIndexCacheSize(int64(10 << 20))(options)
	assert.Equal(t, int64(10<<20), options.indexCacheSize)
}
