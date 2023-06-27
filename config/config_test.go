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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var envVarsDifferent = map[string]string{
	"DEFRA_DATASTORE_STORE":       "memory",
	"DEFRA_DATASTORE_BADGER_PATH": "defra_data",
	"DEFRA_API_ADDRESS":           "localhost:9999",
	"DEFRA_NET_P2PDISABLED":       "true",
	"DEFRA_NET_P2PADDRESS":        "/ip4/0.0.0.0/tcp/9876",
	"DEFRA_NET_RPCADDRESS":        "localhost:7777",
	"DEFRA_NET_RPCTIMEOUT":        "90s",
	"DEFRA_NET_PUBSUB":            "false",
	"DEFRA_NET_RELAY":             "false",
	"DEFRA_LOG_LEVEL":             "error",
	"DEFRA_LOG_STACKTRACE":        "true",
	"DEFRA_LOG_FORMAT":            "json",
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
	"DEFRA_LOG_LEVEL":             "^=+()&**()*(&))",
	"DEFRA_LOG_STACKTRACE":        "^=+()&**()*(&))",
	"DEFRA_LOG_FORMAT":            "^=+()&**()*(&))",
}

func FixtureEnvKeyValue(t *testing.T, key, value string) {
	t.Helper()
	os.Setenv(key, value)
	t.Cleanup(func() {
		os.Unsetenv(key)
	})
}

func FixtureEnvVars(t *testing.T, envVars map[string]string) {
	t.Helper()
	for k, v := range envVars {
		os.Setenv(k, v)
	}
	t.Cleanup(func() {
		for k := range envVars {
			os.Unsetenv(k)
		}
	})
}

func TestConfigValidateBasic(t *testing.T) {
	cfg := DefaultConfig()
	assert.NoError(t, cfg.validate())

	err := cfg.validate()
	assert.NoError(t, err)
	// asserting equality of some unlikely-to-change default values
	assert.Equal(t, "stderr", cfg.Log.Output)
	assert.Equal(t, "csv", cfg.Log.Format)
	assert.Equal(t, false, cfg.API.TLS)
	assert.Equal(t, false, cfg.Net.RelayEnabled)
}

func TestLoadIncorrectValuesFromConfigFile(t *testing.T) {
	var cfg *Config

	testcases := []struct {
		setter func()
		err    error
	}{
		{
			setter: func() {
				cfg.Datastore.Store = "antibadger"
			},
			err: ErrInvalidDatastoreType,
		},
		{
			setter: func() {
				cfg.Log.Level = "antilevel"
			},
			err: ErrInvalidLogLevel,
		},
		{
			setter: func() {
				cfg.Log.Format = "antiformat"
			},

			err: ErrInvalidLogFormat,
		},
	}

	for _, tc := range testcases {
		cfg = DefaultConfig()
		err := cfg.setRootdir(t.TempDir())
		assert.NoError(t, err)
		tc.setter()
		err = cfg.WriteConfigFile()
		assert.NoError(t, err)
		err = cfg.LoadWithRootdir(true)
		assert.ErrorIs(t, err, tc.err)
	}
}

func TestJSONSerialization(t *testing.T) {
	cfg := DefaultConfig()
	var m map[string]any

	b, errSerialize := cfg.ToJSON()
	errUnmarshal := json.Unmarshal(b, &m)

	assert.NoError(t, errUnmarshal)
	assert.NoError(t, errSerialize)
	for k, v := range m {
		if k != "Rootdir" { // Rootdir is not serialized
			assert.NotEmpty(t, v)
		}
	}
}

func TestLoadValidationDefaultsConfigFileEnv(t *testing.T) {
	tmpdir := t.TempDir()
	cfg := DefaultConfig()
	err := cfg.setRootdir(tmpdir)
	assert.NoError(t, err)
	FixtureEnvVars(t, envVarsDifferent)
	errWriteConfig := cfg.WriteConfigFile()

	errLoad := cfg.LoadWithRootdir(true)

	assert.NoError(t, errWriteConfig)
	assert.NoError(t, errLoad)
	assert.Equal(t, "localhost:9999", cfg.API.Address)
	assert.Equal(t, filepath.Join(tmpdir, "defra_data"), cfg.Datastore.Badger.Path)
}

func TestLoadDefaultsEnv(t *testing.T) {
	cfg := DefaultConfig()
	FixtureEnvVars(t, envVarsDifferent)

	err := cfg.LoadWithRootdir(false)

	assert.NoError(t, err)
	assert.Equal(t, "localhost:9999", cfg.API.Address)
	assert.Equal(t, filepath.Join(cfg.Rootdir, "defra_data"), cfg.Datastore.Badger.Path)
}

func TestEnvVariablesAllConsidered(t *testing.T) {
	cfg := DefaultConfig()
	FixtureEnvVars(t, envVarsDifferent)

	err := cfg.LoadWithRootdir(false)

	assert.NoError(t, err)
	assert.Equal(t, "localhost:9999", cfg.API.Address)
	assert.Equal(t, filepath.Join(cfg.Rootdir, "defra_data"), cfg.Datastore.Badger.Path)
	assert.Equal(t, "memory", cfg.Datastore.Store)
	assert.Equal(t, true, cfg.Net.P2PDisabled)
	assert.Equal(t, "/ip4/0.0.0.0/tcp/9876", cfg.Net.P2PAddress)
	assert.Equal(t, "localhost:7777", cfg.Net.RPCAddress)
	assert.Equal(t, "90s", cfg.Net.RPCTimeout)
	assert.Equal(t, false, cfg.Net.PubSubEnabled)
	assert.Equal(t, false, cfg.Net.RelayEnabled)
	assert.Equal(t, "error", cfg.Log.Level)
	assert.Equal(t, true, cfg.Log.Stacktrace)
	assert.Equal(t, "json", cfg.Log.Format)
}

func TestLoadNonExistingConfigFile(t *testing.T) {
	cfg := DefaultConfig()
	err := cfg.setRootdir(t.TempDir())
	assert.NoError(t, err)
	err = cfg.LoadWithRootdir(true)
	assert.ErrorIs(t, err, ErrReadingConfigFile)
}

func TestLoadInvalidConfigFile(t *testing.T) {
	cfg := DefaultConfig()
	tmpdir := t.TempDir()

	errWrite := os.WriteFile(
		filepath.Join(tmpdir, DefaultConfigFileName),
		[]byte("{"),
		0644,
	)
	assert.NoError(t, errWrite)

	err := cfg.setRootdir(tmpdir)
	assert.NoError(t, err)
	errLoad := cfg.LoadWithRootdir(true)
	assert.ErrorIs(t, errLoad, ErrReadingConfigFile)
}

func TestInvalidEnvVars(t *testing.T) {
	cfg := DefaultConfig()
	FixtureEnvVars(t, envVarsInvalid)

	err := cfg.LoadWithRootdir(false)

	assert.ErrorIs(t, err, ErrLoadingConfig)
}

func TestCreateAndLoadCustomConfig(t *testing.T) {
	testdir := t.TempDir()

	cfg := DefaultConfig()
	err := cfg.setRootdir(testdir)
	assert.NoError(t, err)
	// a few valid but non-default changes
	cfg.Net.PubSubEnabled = false
	cfg.Log.Level = "fatal"

	err = cfg.CreateRootDirAndConfigFile()
	assert.NoError(t, err)

	assert.True(t, cfg.ConfigFileExists())

	// check that the config file loads properly
	cfg2 := DefaultConfig()
	err = cfg2.setRootdir(testdir)
	assert.NoError(t, err)
	err = cfg2.LoadWithRootdir(true)
	assert.NoError(t, err)
	assert.Equal(t, cfg.Net.PubSubEnabled, cfg2.Net.PubSubEnabled)
	assert.Equal(t, cfg.Log.Level, cfg2.Log.Level)
}

func TestLoadValidationEnvLoggingConfig(t *testing.T) {
	FixtureEnvKeyValue(t, "DEFRA_LOG_LEVEL", "debug,net=info,log=error,cli=fatal")
	cfg := DefaultConfig()
	err := cfg.LoadWithRootdir(false)
	assert.NoError(t, err)
	assert.Equal(t, "debug", cfg.Log.Level)
	for _, override := range cfg.Log.NamedOverrides {
		switch override.Name {
		case "net":
			assert.Equal(t, "info", override.Level)
		case "log":
			assert.Equal(t, "error", override.Level)
		case "cli":
			assert.Equal(t, "fatal", override.Level)
		default:
			t.Fatal("unexpected named override")
		}
	}
}

func TestLoadValidationEnvLoggerConfig(t *testing.T) {
	FixtureEnvKeyValue(t, "DEFRA_LOG_LOGGER", "net,nocolor=true,level=debug;config,output=stdout,level=info")
	cfg := DefaultConfig()
	err := cfg.LoadWithRootdir(false)
	assert.NoError(t, err)
	for _, override := range cfg.Log.NamedOverrides {
		switch override.Name {
		case "net":
			assert.Equal(t, true, override.NoColor)
			assert.Equal(t, "debug", override.Level)
		case "config":
			assert.Equal(t, "info", override.Level)
			assert.Equal(t, "stdout", override.Output)
		default:
			t.Fatal("unexpected named override")
		}
	}
}

func TestLoadValidationEnvLoggerConfigInvalid(t *testing.T) {
	// logging config parameter not provided as <key>=<value> pair
	FixtureEnvKeyValue(t, "DEFRA_LOG_LOGGER", "net,nocolor,true,level,debug;config,output,stdout,level,info")
	cfg := DefaultConfig()
	err := cfg.LoadWithRootdir(false)
	assert.ErrorIs(t, err, ErrFailedToValidateConfig)

	// invalid logger names
	FixtureEnvKeyValue(t, "DEFRA_LOG_LOGGER", "13;2134;™¡£¡™£∞¡™∞¡™£¢;1234;1")
	cfg = DefaultConfig()
	err = cfg.LoadWithRootdir(false)
	assert.ErrorIs(t, err, ErrFailedToValidateConfig)
}

func TestLoadValidationLoggerConfigFromEnvExhaustive(t *testing.T) {
	FixtureEnvKeyValue(t, "DEFRA_LOG_LOGGER", "net,nocolor=true,level=debug;config,output=stdout,caller=false;logging,stacktrace=true,format=json")
	cfg := DefaultConfig()
	err := cfg.LoadWithRootdir(false)
	assert.NoError(t, err)
	for _, override := range cfg.Log.NamedOverrides {
		switch override.Name {
		case "net":
			assert.Equal(t, true, override.NoColor)
			assert.Equal(t, "debug", override.Level)
		case "config":
			assert.Equal(t, "stdout", override.Output)
			assert.Equal(t, false, override.Caller)
		case "logging":
			assert.Equal(t, true, override.Stacktrace)
			assert.Equal(t, "json", override.Format)
		default:
			t.Fatal("unexpected named override")
		}
	}
}

func TestLoadValidationLoggerConfigFromEnvUnknownParam(t *testing.T) {
	FixtureEnvKeyValue(t, "DEFRA_LOG_LOGGER", "net,unknown=true,level=debug")
	cfg := DefaultConfig()
	err := cfg.LoadWithRootdir(false)
	assert.ErrorIs(t, err, ErrUnknownLoggerParameter)
}

func TestLoadValidationInvalidDatastoreConfig(t *testing.T) {
	FixtureEnvKeyValue(t, "DEFRA_DATASTORE_STORE", "antibadger")
	cfg := DefaultConfig()
	err := cfg.LoadWithRootdir(false)
	assert.ErrorIs(t, err, ErrInvalidDatastoreType)
}

func TestValidationLogger(t *testing.T) {
	testCases := []struct {
		input       string
		expectedErr error
	}{
		{"node,level=debug,output=stdout", nil},
		{"node,level=fatal,format=csv", nil},
		{"node,level=warn", ErrInvalidLogLevel},
		{"node,level=debug;cli,", ErrNotProvidedAsKV},
		{"node,level", ErrNotProvidedAsKV},

		{";", ErrInvalidLoggerConfig},
		{";;", ErrInvalidLoggerConfig},
		{",level=debug", ErrLoggerNameEmpty},
		{"node,bar=baz", ErrUnknownLoggerParameter},            // unknown parameter
		{"m,level=debug,output-json", ErrNotProvidedAsKV},      // key-value pair with invalid separator
		{"myModule,level=debug,extraPart", ErrNotProvidedAsKV}, // additional part after last key-value pair
		{"myModule,=myValue", ErrNotProvidedAsKV},              // empty key
		{",k=v", ErrLoggerNameEmpty},                           // empty module
		{";foo", ErrInvalidLoggerConfig},                       // empty module name
		{"k=v", ErrInvalidLoggerConfig},                        // missing module
		{"debug,net=,log=error,cli=fatal", ErrNotProvidedAsKV}, // empty value

	}

	for _, tc := range testCases {
		cfg := DefaultConfig()
		cfg.Log.Logger = tc.input
		t.Log(tc.input)
		err := cfg.validate()
		assert.ErrorIs(t, err, tc.expectedErr)
	}
}

func TestValidationInvalidEmptyAPIAddress(t *testing.T) {
	cfg := DefaultConfig()
	cfg.API.Address = ""
	err := cfg.validate()
	assert.ErrorIs(t, err, ErrInvalidDatabaseURL)
}

func TestValidationNetConfigPeers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.Peers = "/ip4/127.0.0.1/udp/1234,/ip4/7.7.7.7/tcp/4242/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N"
	err := cfg.validate()
	assert.NoError(t, err)
}

func TestValidationInvalidNetConfigPeers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.Peers = "&(*^(*&^(*&^(*&^))), mmmmh,123123"
	err := cfg.validate()
	assert.ErrorIs(t, err, ErrFailedToValidateConfig)
}

func TestValidationInvalidRPCMaxConnectionIdle(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCMaxConnectionIdle = "123123"
	err := cfg.validate()
	assert.ErrorIs(t, err, ErrFailedToValidateConfig)
}

func TestValidationInvalidRPCTimeout(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCTimeout = "123123"
	err := cfg.validate()
	assert.ErrorIs(t, err, ErrFailedToValidateConfig)
}

func TestValidationRPCTimeoutDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCTimeout = "1s"
	err := cfg.validate()
	assert.NoError(t, err)
}

func TestValidationInvalidRPCTimeoutDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCTimeout = "123123"
	err := cfg.validate()
	assert.ErrorIs(t, err, ErrInvalidRPCTimeout)
}

func TestValidationRPCMaxConnectionIdleDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCMaxConnectionIdle = "1s"
	err := cfg.validate()
	assert.NoError(t, err)
	duration, err := cfg.Net.RPCMaxConnectionIdleDuration()
	assert.NoError(t, err)
	assert.Equal(t, duration, 1*time.Second)
}

func TestValidationInvalidMaxConnectionIdleDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCMaxConnectionIdle = "*ˆ&%*&%"
	err := cfg.validate()
	assert.ErrorIs(t, err, ErrInvalidRPCMaxConnectionIdle)
}

func TestValidationInvalidLoggingConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Log.Level = "546578"
	cfg.Log.Format = "*&)*&"
	err := cfg.validate()
	assert.ErrorIs(t, err, ErrInvalidLogLevel)
}

func TestValidationAddressBasicIncomplete(t *testing.T) {
	cfg := DefaultConfig()
	cfg.API.Address = "localhost"
	err := cfg.validate()
	assert.ErrorIs(t, err, ErrFailedToValidateConfig)
}

func TestValidationAddressLocalhostValid(t *testing.T) {
	cfg := DefaultConfig()
	cfg.API.Address = "localhost:9876"
	err := cfg.validate()
	assert.NoError(t, err)
}

func TestValidationAddress0000Incomplete(t *testing.T) {
	cfg := DefaultConfig()
	cfg.API.Address = "0.0.0.0"
	err := cfg.validate()
	assert.ErrorIs(t, err, ErrFailedToValidateConfig)
}

func TestValidationAddress0000Valid(t *testing.T) {
	cfg := DefaultConfig()
	cfg.API.Address = "0.0.0.0:9876"
	err := cfg.validate()
	assert.NoError(t, err)
}

func TestValidationAddressDomainWithSubdomainValidWithTLSCorrectPortIsInvalid(t *testing.T) {
	cfg := DefaultConfig()
	cfg.API.Address = "sub.example.com:443"
	cfg.API.TLS = true
	err := cfg.validate()
	assert.ErrorIs(t, err, ErrNoPortWithDomain)
}

func TestValidationAddressDomainWithSubdomainWrongPortIsInvalid(t *testing.T) {
	cfg := DefaultConfig()
	cfg.API.Address = "sub.example.com:9876"
	err := cfg.validate()
	assert.ErrorIs(t, err, ErrNoPortWithDomain)
}
