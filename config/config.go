// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package config provides the central point for DefraDB's configuration and related facilities.

[Config] embeds component-specific config structs. Each config struct can have a function providing
default options, a method providing test configurations, a method for validation, a method handling deprecated fields
(e.g. with warnings). This is extensible.

The 'root directory' is where the configuration file and data of a DefraDB instance exists. It is specified as a global
flag `defradb --rootdir path/to/somewhere`, or with the DEFRA_ROOT environment variable.

Some packages of DefraDB provide their own configuration approach (logging, node). For each, a way to go from top-level
configuration to package-specific configuration is provided.

Parameters are determined by, in order of least importance: defaults, configuration file, env. variables, and then CLI
flags. That is, CLI flags can override everything else.

For example `DEFRA_DATASTORE_BADGER_PATH` matches [Config.Datastore.Badger.Path] and in the config file:

	datastore:
		badger:
			path: /tmp/badger

This implementation does not support online modification of configuration.

How to use, e.g. without using a rootdir:

	cfg := config.DefaultConfig()
	cfg.NetConfig.P2PDisabled = true  // as example
	err := cfg.LoadWithoutRootDir()
	if err != nil {
		...

*/
package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/mitchellh/mapstructure"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/viper"

	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v3"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/sourcenetwork/defradb/node"
)

var log = logging.MustNewLogger("defra.config")

const (
	DefraEnvPrefix        = "DEFRA"
	defaultDefraDBRootDir = ".defradb"
	logLevelDebug         = "debug"
	logLevelInfo          = "info"
	logLevelError         = "error"
	logLevelFatal         = "fatal"
)

// Config is DefraDB's main configuration struct, embedding component-specific config structs.
type Config struct {
	Datastore *DatastoreConfig
	API       *APIConfig
	Net       *NetConfig
	Log       *LoggingConfig
}

// Load Config and handles parameters from config file, environment variables.
// To use on a Config struct already loaded with default values from DefaultConfig().
func (cfg *Config) Load(rootDirPath string) error {
	viper.SetConfigName(DefaultDefraDBConfigFileName)
	viper.SetConfigType(configType)
	viper.AddConfigPath(rootDirPath)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	viper.SetEnvPrefix(DefraEnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.Unmarshal(cfg, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc()))
	if err != nil {
		return err
	}

	cfg.handleParams(rootDirPath)
	err = cfg.validate()
	if err != nil {
		return err
	}
	return nil
}

// LoadWithoutRootDir loads Config and handles parameters from defaults, environment variables, and CLI flags -
// not from config file.
// To use on a Config struct already loaded with default values from DefaultConfig().
func (cfg *Config) LoadWithoutRootDir() error {
	// With Viper, we use a config file to provide a basic structure and set defaults, for env. variables to load.
	viper.SetConfigType(configType)
	configbytes, err := cfg.toBytes()
	if err != nil {
		return err
	}
	err = viper.ReadConfig(bytes.NewReader(configbytes))
	if err != nil {
		return err
	}

	viper.SetEnvPrefix(DefraEnvPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err = viper.Unmarshal(cfg, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc()))
	if err != nil {
		return err
	}
	rootDir, err := DefaultRootDir()
	if err != nil {
		log.FatalE(context.Background(), "Could not get home directory", err)
	}

	cfg.handleParams(rootDir)
	err = cfg.validate()
	if err != nil {
		return err
	}
	return nil
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Datastore: defaultDatastoreConfig(),
		API:       defaultAPIConfig(),
		Net:       defaultNetConfig(),
		Log:       defaultLogConfig(),
	}
}

func (cfg *Config) validate() error {
	if err := cfg.Datastore.validate(); err != nil {
		return errors.Wrap("failed to validate Datastore config", err)
	}
	if err := cfg.API.validate(); err != nil {
		return errors.Wrap("failed to validate API config", err)
	}
	if err := cfg.Net.validate(); err != nil {
		return errors.Wrap("failed to validate Net config", err)
	}
	if err := cfg.Log.validate(); err != nil {
		return errors.Wrap("failed to validate Log config", err)
	}
	return nil
}

func (cfg *Config) handleParams(rootDir string) {
	// We prefer using absolute paths.
	if !filepath.IsAbs(cfg.Datastore.Badger.Path) {
		cfg.Datastore.Badger.Path = filepath.Join(rootDir, cfg.Datastore.Badger.Path)
	}
	cfg.setBadgerVLogMaxSize()
}

func (cfg *Config) setBadgerVLogMaxSize() {
	cfg.Datastore.Badger.Options.ValueLogFileSize = int64(cfg.Datastore.Badger.ValueLogFileSize)
}

// DatastoreConfig configures datastores.
type DatastoreConfig struct {
	Store  string
	Memory MemoryConfig
	Badger BadgerConfig
}

// BadgerConfig configures Badger's on-disk / filesystem mode.
type BadgerConfig struct {
	Path             string
	ValueLogFileSize ByteSize
	*badgerds.Options
}

type ByteSize uint64

const (
	B   ByteSize = 1
	KiB          = B << 10
	MiB          = KiB << 10
	GiB          = MiB << 10
	TiB          = GiB << 10
	PiB          = TiB << 10
)

// UnmarshalText calls Set on ByteSize with the given text
func (bs *ByteSize) UnmarshalText(text []byte) error {
	return bs.Set(string(text))
}

// Set parses a string into ByteSize
func (bs *ByteSize) Set(s string) error {
	digitString := ""
	unit := ""
	for _, char := range s {
		if unicode.IsDigit(char) {
			digitString += string(char)
		} else {
			unit += string(char)
		}
	}
	digits, err := strconv.Atoi(digitString)
	if err != nil {
		return err
	}

	switch strings.ToUpper(strings.Trim(unit, " ")) {
	case "B":
		*bs = ByteSize(digits) * B
	case "KB", "KIB":
		*bs = ByteSize(digits) * KiB
	case "MB", "MIB":
		*bs = ByteSize(digits) * MiB
	case "GB", "GIB":
		*bs = ByteSize(digits) * GiB
	case "TB", "TIB":
		*bs = ByteSize(digits) * TiB
	case "PB", "PIB":
		*bs = ByteSize(digits) * PiB
	default:
		*bs = ByteSize(digits)
	}

	return nil
}

// String returns the string formatted output of ByteSize
func (bs *ByteSize) String() string {
	const unit = 1024
	bsInt := int64(*bs)
	if bsInt < unit {
		return fmt.Sprintf("%d", bsInt)
	}
	div, exp := int64(unit), 0
	for n := bsInt / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%d%ciB", bsInt/div, "KMGTP"[exp])
}

// Type returns the type as a string.
func (bs *ByteSize) Type() string {
	return "ByteSize"
}

// MemoryConfig configures of Badger's memory mode.
type MemoryConfig struct {
	Size uint64
}

func defaultDatastoreConfig() *DatastoreConfig {
	// create a copy of the default badger options
	opts := badgerds.DefaultOptions
	return &DatastoreConfig{
		Store: "badger",
		Badger: BadgerConfig{
			Path:             "data",
			ValueLogFileSize: 1 * GiB,
			Options:          &opts,
		},
	}
}

func (dbcfg DatastoreConfig) validate() error {
	switch dbcfg.Store {
	case "badger", "memory":
	default:
		return errors.New(fmt.Sprintf("invalid store type: %s", dbcfg.Store))
	}
	return nil
}

// APIConfig configures the API endpoints.
type APIConfig struct {
	Address string
	TLS     bool
	PubKey  string
	PrivKey string
	Email   string
}

func defaultAPIConfig() *APIConfig {
	return &APIConfig{
		Address: "localhost:9181",
		TLS:     false,
		PubKey:  "certs/server.key",
		PrivKey: "certs/server.crt",
		Email:   "example@example.com",
	}
}

func (apicfg *APIConfig) validate() error {
	if apicfg.Address == "" {
		return errors.New("no database URL provided")
	}
	ip := net.ParseIP(apicfg.Address)
	if strings.HasPrefix(apicfg.Address, "localhost") || strings.HasPrefix(apicfg.Address, ":") || ip != nil {
		_, err := net.ResolveTCPAddr("tcp", apicfg.Address)
		if err != nil {
			return errors.Wrap("invalid database URL", err)
		}
	}
	return nil
}

// AddressToURL provides the API address as URL.
func (apicfg *APIConfig) AddressToURL() string {
	if apicfg.TLS {
		return fmt.Sprintf("https://%s", apicfg.Address)
	}
	return fmt.Sprintf("http://%s", apicfg.Address)
}

// NetConfig configures aspects of network and peer-to-peer.
type NetConfig struct {
	P2PAddress           string
	P2PDisabled          bool
	Peers                string
	PubSubEnabled        bool `mapstructure:"pubsub"`
	RelayEnabled         bool `mapstructure:"relay"`
	RPCAddress           string
	RPCMaxConnectionIdle string
	RPCTimeout           string
	TCPAddress           string
}

func defaultNetConfig() *NetConfig {
	return &NetConfig{
		P2PAddress:           "/ip4/0.0.0.0/tcp/9171",
		P2PDisabled:          false,
		Peers:                "",
		PubSubEnabled:        true,
		RelayEnabled:         false,
		RPCAddress:           "0.0.0.0:9161",
		RPCMaxConnectionIdle: "5m",
		RPCTimeout:           "10s",
		TCPAddress:           "/ip4/0.0.0.0/tcp/9161",
	}
}

func (netcfg *NetConfig) validate() error {
	_, err := time.ParseDuration(netcfg.RPCTimeout)
	if err != nil {
		return errors.New(fmt.Sprintf("invalid RPC timeout: %s", netcfg.RPCTimeout))
	}
	_, err = time.ParseDuration(netcfg.RPCMaxConnectionIdle)
	if err != nil {
		return errors.New(fmt.Sprintf("invalid RPC MaxConnectionIdle: %s", netcfg.RPCMaxConnectionIdle))
	}
	_, err = ma.NewMultiaddr(netcfg.P2PAddress)
	if err != nil {
		return errors.New(fmt.Sprintf("invalid P2P address: %s", netcfg.P2PAddress))
	}
	_, err = net.ResolveTCPAddr("tcp", netcfg.RPCAddress)
	if err != nil {
		return errors.Wrap("invalid RPC address", err)
	}
	if len(netcfg.Peers) > 0 {
		peers := strings.Split(netcfg.Peers, ",")
		maddrs := make([]ma.Multiaddr, len(peers))
		for i, addr := range peers {
			maddrs[i], err = ma.NewMultiaddr(addr)
			if err != nil {
				return errors.New(fmt.Sprintf("failed to parse bootstrap peers: %s", netcfg.Peers))
			}
		}
	}
	return nil
}

// RPCTimeoutDuration gives the RPC timeout as a time.Duration.
func (netcfg *NetConfig) RPCTimeoutDuration() (time.Duration, error) {
	d, err := time.ParseDuration(netcfg.RPCTimeout)
	if err != nil {
		return d, err
	}
	return d, nil
}

// RPCMaxConnectionIdleDuration gives the RPC MaxConnectionIdle as a time.Duration.
func (netcfg *NetConfig) RPCMaxConnectionIdleDuration() (time.Duration, error) {
	d, err := time.ParseDuration(netcfg.RPCMaxConnectionIdle)
	if err != nil {
		return d, err
	}
	return d, nil
}

// NodeConfig provides the Node-specific configuration, from the top-level Net config.
func (cfg *Config) NodeConfig() node.NodeOpt {
	return func(opt *node.Options) error {
		var err error
		err = node.ListenP2PAddrStrings(cfg.Net.P2PAddress)(opt)
		if err != nil {
			return err
		}
		err = node.ListenTCPAddrString(cfg.Net.TCPAddress)(opt)
		if err != nil {
			return err
		}
		opt.EnableRelay = cfg.Net.RelayEnabled
		opt.EnablePubSub = cfg.Net.PubSubEnabled
		opt.DataPath = cfg.Datastore.Badger.Path
		opt.ConnManager, err = node.NewConnManager(100, 400, time.Second*20)
		if err != nil {
			return err
		}
		return nil
	}
}

// LogConfig configures output and logger.
type LoggingConfig struct {
	Level          string
	Stacktrace     bool
	Format         string
	Output         string // logging actually supports multiple output paths, but here only one is supported
	Caller         bool
	NoColor        bool
	NamedOverrides map[string]*NamedLoggingConfig
}

type NamedLoggingConfig struct {
	LoggingConfig
	Name string
}

func defaultLogConfig() *LoggingConfig {
	return &LoggingConfig{
		Level:          logLevelInfo,
		Stacktrace:     false,
		Format:         "csv",
		Output:         "stderr",
		Caller:         false,
		NoColor:        false,
		NamedOverrides: make(map[string]*NamedLoggingConfig),
	}
}

func (logcfg *LoggingConfig) validate() error {
	return nil
}

func (logcfg LoggingConfig) ToLoggerConfig() (logging.Config, error) {
	var loglvl logging.LogLevel
	switch logcfg.Level {
	case logLevelDebug:
		loglvl = logging.Debug
	case logLevelInfo:
		loglvl = logging.Info
	case logLevelError:
		loglvl = logging.Error
	case logLevelFatal:
		loglvl = logging.Fatal
	default:
		return logging.Config{}, errors.New(fmt.Sprintf("invalid log level: %s", logcfg.Level))
	}
	var encfmt logging.EncoderFormat
	switch logcfg.Format {
	case "json":
		encfmt = logging.JSON
	case "csv":
		encfmt = logging.CSV
	default:
		return logging.Config{}, errors.New(fmt.Sprintf("invalid log format: %s", logcfg.Format))
	}
	// handle named overrides
	overrides := make(map[string]logging.Config)
	for name, cfg := range logcfg.NamedOverrides {
		c, err := cfg.ToLoggerConfig()
		if err != nil {
			return logging.Config{}, errors.Wrap("couldn't convert override config", err)
		}
		overrides[name] = c
	}
	return logging.Config{
		Level:                 logging.NewLogLevelOption(loglvl),
		EnableStackTrace:      logging.NewEnableStackTraceOption(logcfg.Stacktrace),
		DisableColor:          logging.NewDisableColorOption(logcfg.NoColor),
		EncoderFormat:         logging.NewEncoderFormatOption(encfmt),
		OutputPaths:           []string{logcfg.Output},
		EnableCaller:          logging.NewEnableCallerOption(logcfg.Caller),
		OverridesByLoggerName: overrides,
	}, nil
}

// this is a copy that doesn't deep copy the NamedOverrides map
// copy is handled by runtime "pass-by-value"
func (logcfg LoggingConfig) copy() LoggingConfig {
	logcfg.NamedOverrides = make(map[string]*NamedLoggingConfig)
	return logcfg
}

func (logcfg *LoggingConfig) GetOrCreateNamedLogger(name string) (*NamedLoggingConfig, error) {
	if name == "" {
		return nil, errors.New("provided name can't be empty for named config")
	}
	if namedCfg, exists := logcfg.NamedOverrides[name]; exists {
		return namedCfg, nil
	}
	// create default and save to overrides
	namedCfg := &NamedLoggingConfig{
		Name:          name,
		LoggingConfig: logcfg.copy(),
	}
	logcfg.NamedOverrides[name] = namedCfg

	return namedCfg, nil
}

// GetLoggingConfig provides logging-specific configuration, from top-level Config.
func (cfg *Config) GetLoggingConfig() (logging.Config, error) {
	return cfg.Log.ToLoggerConfig()
}

// ToJSON serializes the config to a JSON string.
func (c *Config) ToJSON() ([]byte, error) {
	jsonbytes, err := json.Marshal(c)
	if err != nil {
		return []byte{}, errors.Wrap("failed to marshal Config to JSON", err)
	}
	return jsonbytes, nil
}

func (c *Config) toBytes() ([]byte, error) {
	var buffer bytes.Buffer
	tmpl := template.New("configTemplate")
	configTemplate, err := tmpl.Parse(defaultConfigTemplate)
	if err != nil {
		return nil, errors.Wrap("could not parse config template", err)
	}
	if err := configTemplate.Execute(&buffer, c); err != nil {
		return nil, errors.Wrap("could not execute config template", err)
	}
	return buffer.Bytes(), nil
}
