// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keyring

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func TestFileKeyring(t *testing.T) {
	prompt := PromptFunc(func(s string) ([]byte, error) {
		return []byte("secret"), nil
	})

	kr, err := OpenFileKeyring(t.TempDir(), prompt)
	require.NoError(t, err)

	err = kr.Set("peer_key", []byte("abc"))
	require.NoError(t, err)

	// password should be remembered
	assert.Equal(t, []byte("secret"), kr.password)

	err = kr.Set("node_key", []byte("123"))
	require.NoError(t, err)

	peerKey, err := kr.Get("peer_key")
	require.NoError(t, err)
	assert.Equal(t, []byte("abc"), peerKey)

	nodeKey, err := kr.Get("node_key")
	require.NoError(t, err)
	assert.Equal(t, []byte("123"), nodeKey)

	err = kr.Delete("node_key")
	require.NoError(t, err)

	_, err = kr.Get("node_key")
	assert.ErrorIs(t, err, keyring.ErrNotFound)
}
