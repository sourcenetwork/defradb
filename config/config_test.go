// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcenetwork/defradb/logging"
	"github.com/stretchr/testify/assert"
)

var envVarsDifferentThanDefault = map[string]string{
	"DEFRA_DATASTORE_STORE":       "memory",
	"DEFRA_DATASTORE_BADGER_PATH": "defra_data",
	"DEFRA_API_ADDRESS":           "localhost:9999",
	"DEFRA_NET_P2PDISABLED":       "true",
	"DEFRA_NET_P2PADDRESS":        "localhost:9876",
	"DEFRA_NET_RPCADDRESS":        "localhost:7777",
	"DEFRA_NET_RPCTIMEOUT":        "90s",
	"DEFRA_NET_PUBSUB":            "false",
	"DEFRA_NET_RELAY":             "false",
	"DEFRA_LOGGING_LEVEL":         "warn",
	"DEFRA_LOGGING_STACKTRACE":    "false",
	"DEFRA_LOGGING_FORMAT":        "json",
}

var envVarsInvalid = map[string]string{
	"DEFRA_DATASTORE_STORE":       "^=+()&**()*(&))",
	"DEFRA_DATASTORE_BADGER_PATH": "^=+()&**()*(&))",
	"DEFRA_API_ADDRESS":           "^=+()&**()*(&))",
	"DEFRA_NET_P2PDISABLED":       "^=+()&**()*(&))",
	"DEFRA_NET_P2PADDRESS":        "^=+()&**()*(&))",
	"DEFRA_NET_RPCADDRESS":        "^=+()&**()*(&))",
	"DEFRA_NET_RPCTIMEOUT":        "^=+()&**()*(&))",
	"DEFRA_NET_PUBSUB":            "^=+()&**()*(&))",
	"DEFRA_NET_RELAY":             "^=+()&**()*(&))",
	"DEFRA_LOGGING_LEVEL":         "^=+()&**()*(&))",
	"DEFRA_LOGGING_STACKTRACE":    "^=+()&**()*(&))",
	"DEFRA_LOGGING_FORMAT":        "^=+()&**()*(&))",
}

func FixtureEnvVars(envVars map[string]string) {
	for k, v := range envVars {
		os.Setenv(k, v)
	}
}

func FixtureEnvVarsUnset(envVars map[string]string) {
	for k := range envVars {
		os.Unsetenv(k)
	}
}

// Gives a path to a temporary directory containing a default config file
func FixtureDefaultConfigFile(t *testing.T) string {
	dir := t.TempDir()
	cfg := DefaultConfig()

	cfg.writeConfigFile(dir)
	return dir
}

func TestConfigValidateBasic(t *testing.T) {
	cfg := DefaultConfig()
	assert.NoError(t, cfg.validateBasic())
	// Borked configuration gives out error
	cfg.API.Address = "*%(*&"

	err := cfg.validateBasic()

	assert.Error(t, err)
}

func TestJSONSerialization(t *testing.T) {
	cfg := DefaultConfig()
	var m map[string]interface{}

	b, errSerialize := cfg.ToJSON()
	errUnmarshal := json.Unmarshal(b, &m)

	assert.NoError(t, errUnmarshal)
	assert.NoError(t, errSerialize)
	assert.NotEmpty(t, b)
}

func TestLoadDefaultsConfigFileEnv(t *testing.T) {
	dir := t.TempDir()
	cfg := DefaultConfig()
	errWriteConfig := cfg.WriteConfigFileToRootDir(dir)
	FixtureEnvVars(envVarsDifferentThanDefault)
	defer FixtureEnvVarsUnset(envVarsDifferentThanDefault)

	errLoad := cfg.Load(dir)

	assert.NoError(t, errLoad)
	assert.NoError(t, errWriteConfig)
	assert.Equal(t, "localhost:9999", cfg.API.Address)
	assert.Equal(t, filepath.Join(dir, "defra_data"), cfg.Datastore.Badger.Path)
}

func TestLoadDefaultsEnv(t *testing.T) {
	cfg := DefaultConfig()
	FixtureEnvVars(envVarsDifferentThanDefault)
	defer FixtureEnvVarsUnset(envVarsDifferentThanDefault)

	err := cfg.LoadWithoutRootDir()

	assert.NoError(t, err)
	assert.Equal(t, "localhost:9999", cfg.API.Address)
	defaultRootDir, _ := DefaultRootDir()
	assert.Equal(t, filepath.Join(defaultRootDir, "defra_data"), cfg.Datastore.Badger.Path)
}

func TestEnvVariablesAllConsidered(t *testing.T) {
	cfg := DefaultConfig()
	FixtureEnvVars(envVarsDifferentThanDefault)
	defer FixtureEnvVarsUnset(envVarsDifferentThanDefault)

	err := cfg.LoadWithoutRootDir()

	assert.NoError(t, err)
	assert.Equal(t, "localhost:9999", cfg.API.Address)
	defaultRootDir, _ := DefaultRootDir()
	assert.Equal(t, filepath.Join(defaultRootDir, "defra_data"), cfg.Datastore.Badger.Path)
	assert.Equal(t, "memory", cfg.Datastore.Store)
	assert.Equal(t, true, cfg.Net.P2PDisabled)
	assert.Equal(t, "localhost:9876", cfg.Net.P2PAddress)
	assert.Equal(t, "localhost:7777", cfg.Net.RPCAddress)
	assert.Equal(t, "90s", cfg.Net.RPCTimeout)
	assert.Equal(t, false, cfg.Net.PubSubEnabled)
	assert.Equal(t, false, cfg.Net.RelayEnabled)
	assert.Equal(t, "warn", cfg.Logging.Level)
	assert.Equal(t, false, cfg.Logging.Stacktrace)
	assert.Equal(t, "json", cfg.Logging.Format)
}

func TestGetRootDirExists(t *testing.T) {
	dir, exists, err := GetRootDir("/tmp/defra_cli/")

	assert.NoError(t, err)
	assert.Equal(t, "/tmp/defra_cli", dir)
	assert.Equal(t, false, exists)
}

func TestGetRootDir(t *testing.T) {
	os.Setenv("DEFRA_ROOT", "/tmp/defra_env/")
	defer os.Unsetenv("DEFRA_ROOT")

	dir, exists, err := GetRootDir("")

	assert.NoError(t, err)
	assert.Equal(t, "/tmp/defra_env", dir)
	assert.Equal(t, false, exists)
}

func TestLoadNonExistingConfigFile(t *testing.T) {
	cfg := DefaultConfig()
	dir := t.TempDir()

	err := cfg.Load(dir)

	assert.Error(t, err)
}

func TestLoadInvalidConfigFile(t *testing.T) {
	cfg := DefaultConfig()
	dir := t.TempDir()
	ioutil.WriteFile(filepath.Join(dir, defaultDefraDBConfigFileName), []byte("{"), 0644)

	err := cfg.Load(dir)

	assert.Error(t, err)
}

func TestInvalidEnvVars(t *testing.T) {
	cfg := DefaultConfig()
	FixtureEnvVars(envVarsInvalid)
	defer FixtureEnvVarsUnset(envVarsInvalid)

	err := cfg.LoadWithoutRootDir()

	assert.Error(t, err)
}

func TestValidNetConfigPeers(t *testing.T) {
	cfg := DefaultConfig()

	cfg.Net.Peers = "/ip4/127.0.0.1/udp/1234,/ip4/7.7.7.7/tcp/4242/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N"
	err := cfg.LoadWithoutRootDir()

	assert.NoError(t, err)
}

func TestInvalidNetConfigPeers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.Peers = "&(*^(*&^(*&^(*&^))), mmmmh,123123"

	err := cfg.LoadWithoutRootDir()

	assert.Error(t, err)
}

func TestInvalidRPCMaxConnectionIdle(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCMaxConnectionIdle = "123123"

	err := cfg.LoadWithoutRootDir()

	assert.Error(t, err)
}

func TestInvalidRPCTimeout(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCTimeout = "123123"
	err := cfg.LoadWithoutRootDir()
	assert.Error(t, err)
}

func TestValidRPCTimeoutDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCTimeout = "1s"

	cfg.LoadWithoutRootDir()
	_, err := cfg.Net.RPCTimeoutDuration()

	assert.NoError(t, err)
	assert.NoError(t, err)
}

func TestInvalidRPCTimeoutDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCTimeout = "123123"
	cfg.LoadWithoutRootDir()
	_, err := cfg.Net.RPCTimeoutDuration()
	assert.Error(t, err)
}

func TestValidRPCMaxConnectionIdleDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCMaxConnectionIdle = "1s"

	cfg.LoadWithoutRootDir()
	_, err := cfg.Net.RPCMaxConnectionIdleDuration()

	assert.NoError(t, err)
}

func TestInvalidMaxConnectionIdleDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCMaxConnectionIdle = "*ˆ&%*&%"

	cfg.LoadWithoutRootDir()
	_, err := cfg.Net.RPCMaxConnectionIdleDuration()

	assert.Error(t, err)
}

func TestGetLoggingConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Logging.Level = "debug"
	cfg.Logging.Format = "json"
	cfg.Logging.Stacktrace = true
	cfg.Logging.OutputPath = "stdout"

	loggingConfig, err := cfg.GetLoggingConfig()

	assert.NoError(t, err)
	assert.Equal(t, logging.Debug, loggingConfig.Level.LogLevel)
	assert.Equal(t, logging.JSON, loggingConfig.EncoderFormat.EncoderFormat)
	assert.Equal(t, true, loggingConfig.EnableStackTrace.EnableStackTrace)
	assert.Equal(t, "stdout", loggingConfig.OutputPaths[0])
}

func TestInvalidGetLoggingConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Logging.Level = "546578"
	cfg.Logging.Format = "*&)*&"

	cfg.LoadWithoutRootDir()
	_, err := cfg.GetLoggingConfig()

	assert.Error(t, err)
}
