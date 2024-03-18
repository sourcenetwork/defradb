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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCommand(t *testing.T) {
	rootdir := t.TempDir()

	cmd := MakeRootCommand()
	cmd.AddCommand(MakeInitCommand())
	cmd.SetArgs([]string{"init", "--rootdir", rootdir})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.DirExists(t, rootdir)
	assert.FileExists(t, filepath.Join(rootdir, "config.yaml"))
}

func TestInitCommandWithEncryption(t *testing.T) {
	rootdir := t.TempDir()

	cmd := MakeRootCommand()
	cmd.AddCommand(MakeInitCommand())
	cmd.SetArgs([]string{"init", "--rootdir", rootdir, "-e"})

	err := cmd.Execute()
	require.NoError(t, err)

	cfg, err := loadConfig(rootdir, cmd.Root().PersistentFlags())
	require.NoError(t, err)

	assert.Len(t, cfg.GetString("datastore.badger.encryptionkey"), 64)
	assert.NotEqual(t, 0, cfg.GetInt64("datastore.badger.indexcachesize"))
}
