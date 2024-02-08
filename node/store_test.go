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

	"github.com/stretchr/testify/require"
)

func TestNewStoreWithPath(t *testing.T) {
	opts := []StoreOpt{
		WithPath(t.TempDir()),
	}

	store, err := NewStore(opts...)
	require.NoError(t, err)

	err = store.Close()
	require.NoError(t, err)
}

func TestNewStoreWithInMemory(t *testing.T) {
	opts := []StoreOpt{
		WithInMemory(true),
	}

	store, err := NewStore(opts...)
	require.NoError(t, err)

	err = store.Close()
	require.NoError(t, err)
}

func TestNewStoreWithEncryptionKey(t *testing.T) {
	privateKey := make([]byte, 32)
	_, err := rand.Read(privateKey)
	require.NoError(t, err)

	opts := []StoreOpt{
		WithPath(t.TempDir()),
		WithEncryptionKey(privateKey),
	}

	store, err := NewStore(opts...)
	require.NoError(t, err)

	err = store.Close()
	require.NoError(t, err)
}
