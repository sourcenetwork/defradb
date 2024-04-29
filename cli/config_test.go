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

func TestCreateConfig(t *testing.T) {
	rootdir := t.TempDir()
	err := createConfig(rootdir, NewDefraCommand().PersistentFlags())
	require.NoError(t, err)

	// ensure no errors when config already exists
	err = createConfig(rootdir, NewDefraCommand().PersistentFlags())
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(rootdir, "config.yaml"))
}

func TestLoadConfigNotExist(t *testing.T) {
	rootdir := t.TempDir()
	cfg, err := loadConfig(rootdir, NewDefraCommand().PersistentFlags())
	require.NoError(t, err)

	assert.Equal(t, 5, cfg.GetInt("datastore.maxtxnretries"))
	assert.Equal(t, filepath.Join(rootdir, "data"), cfg.GetString("datastore.badger.path"))
	assert.Equal(t, 1<<30, cfg.GetInt("datastore.badger.valuelogfilesize"))
	assert.Equal(t, "badger", cfg.GetString("datastore.store"))
	assert.Equal(t, false, cfg.GetBool("datastore.encryptionDisabled"))

	assert.Equal(t, "127.0.0.1:9181", cfg.GetString("api.address"))
	assert.Equal(t, []string{}, cfg.GetStringSlice("api.allowed-origins"))
	assert.Equal(t, "", cfg.GetString("api.pubkeypath"))
	assert.Equal(t, "", cfg.GetString("api.privkeypath"))

	assert.Equal(t, false, cfg.GetBool("net.p2pdisabled"))
	assert.Equal(t, []string{"/ip4/127.0.0.1/tcp/9171"}, cfg.GetStringSlice("net.p2paddresses"))
	assert.Equal(t, true, cfg.GetBool("net.pubsubenabled"))
	assert.Equal(t, false, cfg.GetBool("net.relay"))
	assert.Equal(t, []string{}, cfg.GetStringSlice("net.peers"))

	assert.Equal(t, "info", cfg.GetString("log.level"))
	assert.Equal(t, "stderr", cfg.GetString("log.output"))
	assert.Equal(t, "text", cfg.GetString("log.format"))
	assert.Equal(t, false, cfg.GetBool("log.stacktrace"))
	assert.Equal(t, false, cfg.GetBool("log.source"))
	assert.Equal(t, "", cfg.GetString("log.overrides"))
	assert.Equal(t, false, cfg.GetBool("log.nocolor"))

	assert.Equal(t, filepath.Join(rootdir, "keys"), cfg.GetString("keyring.path"))
	assert.Equal(t, false, cfg.GetBool("keyring.disabled"))
}
