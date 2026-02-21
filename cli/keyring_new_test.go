// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyringNew(t *testing.T) {
	rootdir := t.TempDir()
	err := os.Setenv("DEFRA_KEYRING_SECRET", "password")
	require.NoError(t, err)

	cmd := NewDefraCommand(context.Background())
	cmd.SetArgs([]string{"keyring", "new", "--rootdir", rootdir})

	err = cmd.Execute()
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(rootdir, "keys", encryptionKeyName))
	assert.FileExists(t, filepath.Join(rootdir, "keys", peerKeyName))
}

func TestKeyringNewNoEncryptionKey(t *testing.T) {
	rootdir := t.TempDir()
	err := os.Setenv("DEFRA_KEYRING_SECRET", "password")
	require.NoError(t, err)

	cmd := NewDefraCommand(context.Background())
	cmd.SetArgs([]string{"keyring", "new", "--no-encryption", "--rootdir", rootdir})

	err = cmd.Execute()
	require.NoError(t, err)

	assert.NoFileExists(t, filepath.Join(rootdir, "keys", encryptionKeyName))
	assert.FileExists(t, filepath.Join(rootdir, "keys", peerKeyName))
}

func TestKeyringNewNoPeerKey(t *testing.T) {
	rootdir := t.TempDir()
	err := os.Setenv("DEFRA_KEYRING_SECRET", "password")
	require.NoError(t, err)

	cmd := NewDefraCommand(context.Background())
	cmd.SetArgs([]string{"keyring", "new", "--no-peer-key", "--rootdir", rootdir})

	err = cmd.Execute()
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(rootdir, "keys", encryptionKeyName))
	assert.NoFileExists(t, filepath.Join(rootdir, "keys", peerKeyName))
}

func TestKeyringNewOverwrite(t *testing.T) {
	rootdir := t.TempDir()
	err := os.Setenv("DEFRA_KEYRING_SECRET", "password")
	require.NoError(t, err)

	cmd := NewDefraCommand(context.Background())
	cmd.SetArgs([]string{"keyring", "new", "--rootdir", rootdir})

	err = cmd.Execute()
	require.NoError(t, err)

	cmd2 := NewDefraCommand(context.Background())
	cmd2.SetArgs([]string{"keyring", "new", "--rootdir", rootdir})
	err = cmd2.Execute()

	require.Error(t, err)
}

func TestKeyringNewOverwriteForce(t *testing.T) {
	rootdir := t.TempDir()
	err := os.Setenv("DEFRA_KEYRING_SECRET", "password")
	require.NoError(t, err)

	cmd := NewDefraCommand(context.Background())
	cmd.SetArgs([]string{"keyring", "new", "--rootdir", rootdir})

	err = cmd.Execute()
	require.NoError(t, err)

	cmd2 := NewDefraCommand(context.Background())
	cmd2.SetArgs([]string{"keyring", "new", "--rootdir", rootdir, "--force"})
	err = cmd2.Execute()

	require.NoError(t, err)
}
