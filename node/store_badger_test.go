// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js

package node

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithBadgerInMemory(t *testing.T) {
	options := &StoreOptions{}
	WithBadgerInMemory(true)(options)
	assert.Equal(t, true, options.badgerInMemory)
}

func TestWithBadgerFileSize(t *testing.T) {
	options := &StoreOptions{}
	WithBadgerFileSize(int64(5 << 30))(options)
	assert.Equal(t, int64(5<<30), options.badgerFileSize)
}

func TestWithBadgerEncryptionKey(t *testing.T) {
	encryptionKey := make([]byte, 32)
	_, err := rand.Read(encryptionKey)
	require.NoError(t, err)

	options := &StoreOptions{}
	WithBadgerEncryptionKey(encryptionKey)(options)
	assert.Equal(t, encryptionKey, options.badgerEncryptionKey)
}
