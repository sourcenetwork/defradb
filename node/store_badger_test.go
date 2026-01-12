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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client/options"
)

func TestNodeStoreOptions_BadgerInMemory(t *testing.T) {
	storeOpts := options.NodeStore()
	storeOpts.BadgerInMemory = true
	assert.Equal(t, true, storeOpts.BadgerInMemory)
}

func TestNodeStoreOptions_BadgerFileSize(t *testing.T) {
	storeOpts := options.NodeStore()
	storeOpts.BadgerFileSize = int64(5 << 30)
	assert.Equal(t, int64(5<<30), storeOpts.BadgerFileSize)
}

func TestNodeStoreOptions_BadgerEncryptionKey(t *testing.T) {
	encryptionKey := make([]byte, 32)
	_, err := rand.Read(encryptionKey)
	require.NoError(t, err)

	storeOpts := options.NodeStore()
	storeOpts.BadgerEncryptionKey = encryptionKey
	assert.Equal(t, encryptionKey, storeOpts.BadgerEncryptionKey)
}
