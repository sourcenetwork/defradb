package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigNotExist(t *testing.T) {
	rootdir := t.TempDir()
	cfg, err := LoadConfig(rootdir)
	require.NoError(t, err)

	assert.Equal(t, "badger", cfg.GetString("datastore.store"))
	assert.Equal(t, 5, cfg.GetInt("datastore.maxtxnretries"))

	assert.Equal(t, filepath.Join(rootdir, "data"), cfg.GetString("datastore.badger.path"))
	assert.Equal(t, 1073741824, cfg.GetInt("datastore.badger.valuelogfilesize"))

	assert.Equal(t, "localhost:9181", cfg.GetString("api.address"))
	assert.Equal(t, false, cfg.GetBool("api.tls"))
	assert.Equal(t, []string(nil), cfg.GetStringSlice("api.allowed-origins"))
	assert.Equal(t, filepath.Join(rootdir, "cert.pub"), cfg.GetString("api.pubkeypath"))
	assert.Equal(t, filepath.Join(rootdir, "cert.key"), cfg.GetString("api.privkeypath"))
	assert.Equal(t, "example@example.com", cfg.GetString("api.email"))

	assert.Equal(t, false, cfg.GetBool("net.p2pdisabled"))
	assert.Equal(t, []string{"/ip4/127.0.0.1/tcp/9171"}, cfg.GetStringSlice("net.p2paddresses"))
	assert.Equal(t, true, cfg.GetBool("net.pubsubenabled"))
	assert.Equal(t, false, cfg.GetBool("net.relay"))
	assert.Equal(t, []string(nil), cfg.GetStringSlice("net.peers"))

	assert.Equal(t, "info", cfg.GetString("log.level"))
	assert.Equal(t, true, cfg.GetBool("log.stacktrace"))
	assert.Equal(t, "csv", cfg.GetString("log.format"))
	assert.Equal(t, "stderr", cfg.GetString("log.output"))
	assert.Equal(t, false, cfg.GetBool("log.nocolor"))
	assert.Equal(t, false, cfg.GetBool("log.caller"))
}

func TestLoadConfigExists(t *testing.T) {
	rootdir := t.TempDir()
	err := WriteDefaultConfig(rootdir)
	require.NoError(t, err)

	_, err = LoadConfig(rootdir)
	require.NoError(t, err)
}

func TestWriteDefaultConfig(t *testing.T) {
	rootdir := t.TempDir()
	err := WriteDefaultConfig(rootdir)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(rootdir, configName))
	require.NoError(t, err)
	assert.Equal(t, defaultConfig, data)
}
