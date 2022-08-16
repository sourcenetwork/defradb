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
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/node"
	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, cfg.validate())
	// Borked configuration gives out error
	cfg.API.Address = "*%(*&"

	err := cfg.validate()

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

// TODO test more the JSON serialization
// 2022-06-20T14:03:14.284-0500, WARN, defra.cli, WTF initConfig, {"cfg": "eyJEYXRhc3RvcmUiOnsiU3RvcmUiOiJiYWRnZXIiLCJNZW1vcnkiOnsiU2l6ZSI6MH0sIkJhZGdlciI6eyJQYXRoIjoiZGF0YSJ9fSwiQVBJIjp7IkFkZHJlc3MiOiJsb2NhbGhvc3Q6OTE4MSJ9LCJOZXQiOnsiUDJQQWRkcmVzcyI6Ii9pcDQvMC4wLjAuMC90Y3AvOTE3MSIsIlAyUERpc2FibGVkIjpmYWxzZSwiUGVlcnMiOiIiLCJQdWJTdWJFbmFibGVkIjp0cnVlLCJSZWxheUVuYWJsZWQiOnRydWUsIlJQQ0FkZHJlc3MiOiIwLjAuMC4wOjkxNjEiLCJSUENNYXhDb25uZWN0aW9uSWRsZSI6IjVtIiwiUlBDVGltZW91dCI6IjEwcyIsIlRDUEFkZHJlc3MiOiIvaXA0LzAuMC4wLjAvdGNwLzkxNjEifSwiTG9nZ2luZyI6eyJMZXZlbCI6ImRlYnVnIiwiU3RhY2t0cmFjZSI6ZmFsc2UsIkZvcm1hdCI6ImNzdiIsIk91dHB1dFBhdGgiOiJzdGRvdXQiLCJDb2xvciI6dHJ1ZX19"}

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
	assert.Equal(t, "/ip4/0.0.0.0/tcp/9876", cfg.Net.P2PAddress)
	assert.Equal(t, "localhost:7777", cfg.Net.RPCAddress)
	assert.Equal(t, "90s", cfg.Net.RPCTimeout)
	assert.Equal(t, false, cfg.Net.PubSubEnabled)
	assert.Equal(t, false, cfg.Net.RelayEnabled)
	assert.Equal(t, "error", cfg.Log.Level)
	assert.Equal(t, true, cfg.Log.Stacktrace)
	assert.Equal(t, "json", cfg.Log.Format)
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

	errWrite := os.WriteFile(
		filepath.Join(dir, DefaultDefraDBConfigFileName),
		[]byte("{"),
		0644,
	)
	assert.NoError(t, errWrite)

	errLoad := cfg.Load(dir)
	assert.Error(t, errLoad)
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
	cfg.Log.Level = "debug"
	cfg.Log.Format = "json"
	cfg.Log.Stacktrace = true
	cfg.Log.Output = "stdout"

	loggingConfig, err := cfg.GetLoggingConfig()

	assert.NoError(t, err)
	assert.Equal(t, logging.Debug, loggingConfig.Level.LogLevel)
	assert.Equal(t, logging.JSON, loggingConfig.EncoderFormat.EncoderFormat)
	assert.Equal(t, true, loggingConfig.EnableStackTrace.EnableStackTrace)
	assert.Equal(t, "stdout", loggingConfig.OutputPaths[0])
}

func TestInvalidGetLoggingConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Log.Level = "546578"
	cfg.Log.Format = "*&)*&"

	cfg.LoadWithoutRootDir()
	_, err := cfg.GetLoggingConfig()

	assert.Error(t, err)
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
	cfg.LoadWithoutRootDir()

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

func TestUnmarshallByteSize(t *testing.T) {
	var bs ByteSize

	b := []byte("10")
	err := bs.UnmarshalText(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 10*B, bs)

	b = []byte("10B")
	err = bs.UnmarshalText(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 10*B, bs)

	b = []byte("10 B")
	err = bs.UnmarshalText(b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 10*B, bs)

	kb := []byte("10KB")
	err = bs.UnmarshalText(kb)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 10*KB, bs)

	kb = []byte("10 kb")
	err = bs.UnmarshalText(kb)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 10*KB, bs)

	mb := []byte("10MB")
	err = bs.UnmarshalText(mb)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 10*MB, bs)

	gb := []byte("10GB")
	err = bs.UnmarshalText(gb)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 10*GB, bs)

	tb := []byte("10TB")
	err = bs.UnmarshalText(tb)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 10*TB, bs)

	pb := []byte("10PB")
	err = bs.UnmarshalText(pb)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 10*PB, bs)

	eb := []byte("१")
	err = bs.UnmarshalText(eb)
	assert.Error(t, err)
}

func TestByteSizeType(t *testing.T) {
	var bs ByteSize
	assert.Equal(t, "ByteSize", bs.Type())
}

func TestByteSizeToString(t *testing.T) {
	b := 999 * B
	assert.Equal(t, "999", b.String())

	mb := 10 * MB
	assert.Equal(t, "10MB", mb.String())
}
