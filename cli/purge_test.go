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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPurgeCommandWithoutForceFlag_ReturnsError(t *testing.T) {
	rootDir := t.TempDir()

	cmd := NewDefraCommand()
	cmd.SetArgs([]string{"purge", "--rootdir", rootDir})

	err := cmd.Execute()
	require.ErrorIs(t, err, ErrPurgeForceFlagRequired)
}

func TestPurgeCommandWithForceFlag_Succeeds(t *testing.T) {
	rootDir := t.TempDir()

	dataDir := filepath.Join(rootDir, "data")
	err := os.MkdirAll(dataDir, 0755)
	require.NoError(t, err)

	cmd := NewDefraCommand()
	cmd.SetArgs([]string{"purge", "--force", "--rootdir", rootDir})

	err = cmd.Execute()
	require.NoError(t, err)
	assert.NoDirExists(t, dataDir)
}
