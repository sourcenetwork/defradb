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
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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

func TestPurgeCommandWithEmptyDirectory_DoesNothing(t *testing.T) {
	rootDir := t.TempDir()

	cmd := NewDefraCommand()
	cmd.SetArgs([]string{"purge", "--force", "--rootdir", rootDir})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestPurgeCommandWithForceFlag_Succeeds(t *testing.T) {
	rootDir := t.TempDir()

	var dataDir string
	// load the config and create the data directory
	configLoader = func(rootdir string, flags *pflag.FlagSet) (*viper.Viper, error) {
		cfg, err := loadConfig(rootdir, flags)
		if err != nil {
			return nil, err
		}
		dataDir = cfg.GetString("datastore.badger.path")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	cmd := NewDefraCommand()
	cmd.SetArgs([]string{"purge", "--force", "--rootdir", rootDir})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.NoDirExists(t, dataDir)
}
