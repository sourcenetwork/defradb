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

	ma "github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/node"
)

var envVarsDifferentThanDefault = map[string]string{
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
	cfg.Rootdir = dir
	err := cfg.WriteConfigFile()
	assert.NoError(t, err)
	return dir
}

func TestConfigValidateBasic(t *testing.T) {
	cfg := DefaultConfig()
	assert.NoError(t, cfg.validate())
	// Borked configuration gives out error
	cfg.API.Address = "localhost"

	err := cfg.validate()

	assert.ErrorIs(t, err, ErrFailedToValidateConfig)
}

func TestJSONSerialization(t *testing.T) {
	cfg := DefaultConfig()
	var m map[string]any

	b, errSerialize := cfg.ToJSON()
	errUnmarshal := json.Unmarshal(b, &m)

	assert.NoError(t, errUnmarshal)
	assert.NoError(t, errSerialize)
	for _, v := range m {
		assert.NotEmpty(t, v)
	}
}

func TestLoadDefaultsConfigFileEnv(t *testing.T) {
	tmpdir := t.TempDir()
	cfg := DefaultConfig()
	cfg.Rootdir = tmpdir
	FixtureEnvVars(envVarsDifferentThanDefault)
	defer FixtureEnvVarsUnset(envVarsDifferentThanDefault)
	errWriteConfig := cfg.WriteConfigFile()

	errLoad := cfg.LoadWithRootdir(true)

	assert.NoError(t, errWriteConfig)
	assert.NoError(t, errLoad)
	assert.Equal(t, "localhost:9999", cfg.API.Address)
	assert.Equal(t, filepath.Join(tmpdir, "defra_data"), cfg.Datastore.Badger.Path)
}

func TestLoadDefaultsEnv(t *testing.T) {
	cfg := DefaultConfig()
	FixtureEnvVars(envVarsDifferentThanDefault)
	defer FixtureEnvVarsUnset(envVarsDifferentThanDefault)

	err := cfg.LoadWithRootdir(false)

	assert.NoError(t, err)
	assert.Equal(t, "localhost:9999", cfg.API.Address)
	assert.Equal(t, filepath.Join(cfg.Rootdir, "defra_data"), cfg.Datastore.Badger.Path)
}

func TestEnvVariablesAllConsidered(t *testing.T) {
	cfg := DefaultConfig()
	FixtureEnvVars(envVarsDifferentThanDefault)
	defer FixtureEnvVarsUnset(envVarsDifferentThanDefault)

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
	cfg.Rootdir = t.TempDir()
	err := cfg.LoadWithRootdir(true)
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

	cfg.Rootdir = tmpdir
	errLoad := cfg.LoadWithRootdir(true)
	assert.ErrorIs(t, errLoad, ErrReadingConfigFile)
}

func TestInvalidEnvVars(t *testing.T) {
	cfg := DefaultConfig()
	FixtureEnvVars(envVarsInvalid)
	defer FixtureEnvVarsUnset(envVarsInvalid)

	err := cfg.LoadWithRootdir(false)

	assert.ErrorIs(t, err, ErrLoadingConfig)
}

func TestValidNetConfigPeers(t *testing.T) {
	cfg := DefaultConfig()

	cfg.Net.Peers = "/ip4/127.0.0.1/udp/1234,/ip4/7.7.7.7/tcp/4242/p2p/QmYyQSo1c1Ym7orWxLYvCrM2EmxFTANf8wXmmE7DWjhx5N"
	err := cfg.LoadWithRootdir(false)

	assert.NoError(t, err)
}

func TestInvalidNetConfigPeers(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.Peers = "&(*^(*&^(*&^(*&^))), mmmmh,123123"

	err := cfg.LoadWithRootdir(false)

	// The following error is an artefact of the current LoadWithRootdir
	// It should be a validation error
	assert.ErrorIs(t, err, ErrReadingConfigFile)
}

func TestInvalidRPCMaxConnectionIdle(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCMaxConnectionIdle = "123123"

	err := cfg.LoadWithRootdir(false)

	assert.ErrorIs(t, err, ErrFailedToValidateConfig)
}

func TestInvalidRPCTimeout(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCTimeout = "123123"
	err := cfg.LoadWithRootdir(false)
	assert.ErrorIs(t, err, ErrFailedToValidateConfig)
}

func TestValidRPCTimeoutDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCTimeout = "1s"

	err := cfg.LoadWithRootdir(false)
	assert.NoError(t, err)

	_, err = cfg.Net.RPCTimeoutDuration()
	assert.NoError(t, err)
}

func TestInvalidRPCTimeoutDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCTimeout = "123123"

	err := cfg.LoadWithRootdir(false)
	assert.ErrorIs(t, err, ErrInvalidRPCTimeout)

	_, err = cfg.Net.RPCTimeoutDuration()
	assert.ErrorIs(t, err, ErrInvalidRPCTimeout)
}

func TestValidRPCMaxConnectionIdleDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCMaxConnectionIdle = "1s"

	assert.NoError(t, cfg.LoadWithRootdir(false))
	_, err := cfg.Net.RPCMaxConnectionIdleDuration()

	assert.NoError(t, err)
}

func TestInvalidMaxConnectionIdleDuration(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.RPCMaxConnectionIdle = "*ˆ&%*&%"

	err := cfg.LoadWithRootdir(false)
	// The following error is an artefact of the current LoadWithRootdir
	// It should be a validation error
	assert.ErrorIs(t, err, ErrReadingConfigFile)

	_, err = cfg.Net.RPCMaxConnectionIdleDuration()
	assert.ErrorIs(t, err, ErrInvalidRPCMaxConnectionIdle)
}

func TestInvalidGetLoggingConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Log.Level = "546578"
	cfg.Log.Format = "*&)*&"

	err := cfg.LoadWithRootdir(false)
	// The following error is an artefact of the current LoadWithRootdir
	// It should be a validation error
	assert.ErrorIs(t, err, ErrReadingConfigFile)
}

func TestNodeConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Net.P2PAddress = "/ip4/0.0.0.0/tcp/9179"
	cfg.Net.TCPAddress = "/ip4/0.0.0.0/tcp/9169"
	cfg.Net.RPCTimeout = "100s"
	cfg.Net.RPCMaxConnectionIdle = "111s"
	cfg.Net.RelayEnabled = true
	cfg.Net.PubSubEnabled = true
	cfg.Datastore.Badger.Path = "/tmp/defra_cli/badger"
	assert.NoError(t, cfg.LoadWithRootdir(false))

	nodeConfig := cfg.NodeConfig()
	options, errOptionsMerge := node.NewMergedOptions(nodeConfig)

	// confirming it provides the same config as a manually constructed node.Options
	p2pAddr, errP2P := ma.NewMultiaddr(cfg.Net.P2PAddress)
	tcpAddr, errTCP := ma.NewMultiaddr(cfg.Net.TCPAddress)
	connManager, errConnManager := node.NewConnManager(100, 400, time.Second*20)
	expectedOptions := node.Options{
		ListenAddrs:  []ma.Multiaddr{p2pAddr},
		TCPAddr:      tcpAddr,
		DataPath:     "/tmp/defra_cli/badger",
		EnablePubSub: true,
		EnableRelay:  true,
		ConnManager:  connManager,
	}
	assert.NoError(t, errOptionsMerge)
	assert.NoError(t, errP2P)
	assert.NoError(t, errTCP)
	assert.NoError(t, errConnManager)
	for k, v := range options.ListenAddrs {
		assert.Equal(t, expectedOptions.ListenAddrs[k], v)
	}
	assert.Equal(t, expectedOptions.TCPAddr.String(), options.TCPAddr.String())
	assert.Equal(t, expectedOptions.DataPath, options.DataPath)
	assert.Equal(t, expectedOptions.EnablePubSub, options.EnablePubSub)
	assert.Equal(t, expectedOptions.EnableRelay, options.EnableRelay)
}

func TestUnmarshalByteSize(t *testing.T) {
	var bs ByteSize

	b := []byte("10")
	err := bs.UnmarshalText(b)
	assert.NoError(t, err)
	assert.Equal(t, 10*B, bs)

	b = []byte("10B")
	err = bs.UnmarshalText(b)
	assert.NoError(t, err)
	assert.Equal(t, 10*B, bs)

	b = []byte("10 B")
	err = bs.UnmarshalText(b)
	assert.NoError(t, err)
	assert.Equal(t, 10*B, bs)

	kb := []byte("10KB")
	err = bs.UnmarshalText(kb)
	assert.NoError(t, err)
	assert.Equal(t, 10*KiB, bs)

	kb = []byte("10KiB")
	err = bs.UnmarshalText(kb)
	assert.NoError(t, err)
	assert.Equal(t, 10*KiB, bs)

	kb = []byte("10 kb")
	err = bs.UnmarshalText(kb)
	assert.NoError(t, err)
	assert.Equal(t, 10*KiB, bs)

	mb := []byte("10MB")
	err = bs.UnmarshalText(mb)
	assert.NoError(t, err)
	assert.Equal(t, 10*MiB, bs)

	mb = []byte("10MiB")
	err = bs.UnmarshalText(mb)
	assert.NoError(t, err)
	assert.Equal(t, 10*MiB, bs)

	gb := []byte("10GB")
	err = bs.UnmarshalText(gb)
	assert.NoError(t, err)
	assert.Equal(t, 10*GiB, bs)

	gb = []byte("10GiB")
	err = bs.UnmarshalText(gb)
	assert.NoError(t, err)
	assert.Equal(t, 10*GiB, bs)

	tb := []byte("10TB")
	err = bs.UnmarshalText(tb)
	assert.NoError(t, err)
	assert.Equal(t, 10*TiB, bs)

	tb = []byte("10TiB")
	err = bs.UnmarshalText(tb)
	assert.NoError(t, err)
	assert.Equal(t, 10*TiB, bs)

	pb := []byte("10PB")
	err = bs.UnmarshalText(pb)
	assert.NoError(t, err)
	assert.Equal(t, 10*PiB, bs)

	pb = []byte("10PiB")
	err = bs.UnmarshalText(pb)
	assert.NoError(t, err)
	assert.Equal(t, 10*PiB, bs)

	eb := []byte("१")
	err = bs.UnmarshalText(eb)
	assert.ErrorIs(t, err, ErrUnableToParseByteSize)
}

func TestByteSizeType(t *testing.T) {
	var bs ByteSize
	assert.Equal(t, "ByteSize", bs.Type())
}

func TestByteSizeToString(t *testing.T) {
	b := 999 * B
	assert.Equal(t, "999", b.String())

	mb := 10 * MiB
	assert.Equal(t, "10MiB", mb.String())
}

func TestCreateAndLoadCustomConfig(t *testing.T) {
	testdir := t.TempDir()

	cfg := DefaultConfig()
	cfg.Rootdir = testdir
	// a few valid but non-default changes
	cfg.Net.PubSubEnabled = false
	cfg.Log.Level = "fatal"

	err := cfg.CreateRootDirAndConfigFile()
	assert.NoError(t, err)

	assert.True(t, cfg.ConfigFileExists())
	// check that the config file loads properly
	cfg2 := DefaultConfig()
	cfg2.Rootdir = testdir
	err = cfg2.LoadWithRootdir(true)
	assert.NoError(t, err)
	assert.Equal(t, cfg.Net.PubSubEnabled, cfg2.Net.PubSubEnabled)
	assert.Equal(t, cfg.Log.Level, cfg2.Log.Level)
}

// not sure how this behaves in parallel
func envSet(t *testing.T, envs map[string]string) (cleanup func()) {
	originalEnvs := map[string]string{}

	for k, v := range envs {
		if orig, ok := os.LookupEnv(k); ok {
			originalEnvs[k] = orig
		}
		t.Setenv(k, v)
	}

	return func() {
		for k := range envs {
			orig, has := originalEnvs[k]
			if has {
				t.Setenv(k, orig)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}
}

func TestDoNotSupportRootdirFromEnv(t *testing.T) {
	tmpdir := t.TempDir()
	t.Cleanup(envSet(t, map[string]string{
		"DEFRA_ROOTDIR": tmpdir,
	}))
	cfg := DefaultConfig()
	err := cfg.LoadWithRootdir(false)
	assert.Equal(t, cfg.Rootdir, DefaultRootDir())
	assert.NoError(t, err)
}

func TestLoggingConfigFromEnv(t *testing.T) {
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

func TestLoggerConfigFromEnv(t *testing.T) {
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

func TestLoggerConfigFromEnvBroken(t *testing.T) {
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

func TestLoggerConfigFromEnvExhaustive(t *testing.T) {
	FixtureEnvKeyValue(t, "DEFRA_LOG_LOGGER", "net,nocolor=true,level=debug;config,output=stdout,caller=false;logging,stacktrace=true,format=json")
	cfg := DefaultConfig()
	err := cfg.LoadWithRootdir(false)
	assert.NoError(t, err)
}

func TestLoggerConfigFromEnvUnknownParam(t *testing.T) {
	FixtureEnvKeyValue(t, "DEFRA_LOG_LOGGER", "net,unknown=true,level=debug")
	cfg := DefaultConfig()
	err := cfg.LoadWithRootdir(false)
	assert.ErrorIs(t, err, ErrUnknownLoggerParameter)
}

func TestInvalidDatastoreConfig(t *testing.T) {
	FixtureEnvKeyValue(t, "DEFRA_DATASTORE_STORE", "antibadger")
	cfg := DefaultConfig()
	err := cfg.LoadWithRootdir(false)
	assert.ErrorIs(t, err, ErrInvalidDatastoreType)
}
