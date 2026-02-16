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
	"github.com/sourcenetwork/defradb/internal/utils"
)

func TestSetBadgerInMemory(t *testing.T) {
	opts := utils.NewOptions(options.Node().Store().SetBadgerInMemory(true).Node())
	assert.Equal(t, true, opts.Store.BadgerInMemory)
}

func TestSetBadgerFileSize(t *testing.T) {
	opts := utils.NewOptions(options.Node().Store().SetBadgerFileSize(int64(5 << 30)).Node())
	assert.Equal(t, int64(5<<30), opts.Store.BadgerFileSize)
}

func TestSetBadgerEncryptionKey(t *testing.T) {
	encryptionKey := make([]byte, 32)
	_, err := rand.Read(encryptionKey)
	require.NoError(t, err)

	opts := utils.NewOptions(options.Node().Store().SetBadgerEncryptionKey(encryptionKey).Node())
	assert.Equal(t, encryptionKey, opts.Store.BadgerEncryptionKey)
}
