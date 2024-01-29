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
default options, a method providing test configurations, a method for validation, a method handling
deprecated fields (e.g. with warnings). This is extensible.

The 'root directory' is where the configuration file and data of a DefraDB instance exists.
It is specified as a global flag `defradb --rootdir path/to/somewhere.

Some packages of DefraDB provide their own configuration approach (logging, node).
For each, a way to go from top-level configuration to package-specific configuration is provided.

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
	err := cfg.LoadWithRootdir(false)
	if err != nil {
		...
*/
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/mitchellh/mapstructure"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/net/idna"

	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/logging"
)

var log = logging.MustNewLogger("config")

const (
	DefaultAPIEmail = "example@example.com"
	RootdirKey      = "rootdircli"
	defraEnvPrefix  = "DEFRA"
	logLevelDebug   = "debug"
	logLevelInfo    = "info"
	logLevelError   = "error"
	logLevelFatal   = "fatal"
)

// Config is DefraDB's main configuration struct, embedding component-specific config structs.
type Config struct {
	Datastore *DatastoreConfig
	API       *APIConfig
	Net       *NetConfig
	Log       *LoggingConfig
	Rootdir   string
	v         *viper.Viper
}

// DefaultConfig returns the default configuration (or panics).
func DefaultConfig() *Config {
	cfg := &Config{
		Datastore: defaultDatastoreConfig(),
		API:       defaultAPIConfig(),
		Net:       defaultNetConfig(),
		Log:       defaultLogConfig(),
		Rootdir:   "",
		v:         viper.New(),
	}

	cfg.v.SetEnvPrefix(defraEnvPrefix)
	cfg.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg.v.SetConfigName(DefaultConfigFileName)
	cfg.v.SetConfigType(configType)

	cfg.Persist()

	return cfg
}

// Persist persists manually set config parameters to the viper config.
func (cfg *Config) Persist() {
	// Load new values in viper.
	b, err := cfg.toBytes()
	if err != nil {
		panic(err)
	}
	if err = cfg.v.ReadConfig(bytes.NewReader(b)); err != nil {
		panic(NewErrReadingConfigFile(err))
	}
}

// LoadWithRootdir loads a Config with parameters from defaults, config file, environment variables, and CLI flags.
// It loads from config file when `fromFile` is true, otherwise it loads directly from a default configuration.
// Use on a Config struct already loaded with default values from DefaultConfig().
// To be executed once at the beginning of the program.
func (cfg *Config) LoadWithRootdir(withRootdir bool) error {
	// Use default logging configuration here, so that
	// we can log errors in a consistent way even in the case of early failure.
	defaultLogCfg := defaultLogConfig()
	if err := defaultLogCfg.load(); err != nil {
		return err
	}

	if withRootdir {
		if err := cfg.v.ReadInConfig(); err != nil {
			return NewErrReadingConfigFile(err)
		}
	}

	cfg.v.AutomaticEnv()

	if err := cfg.paramsPreprocessing(); err != nil {
		return err
	}
	// We load the viper configuration in the Config struct.
	if err := cfg.v.Unmarshal(cfg, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc())); err != nil {
		return NewErrLoadingConfig(err)
	}
	if err := cfg.validate(); err != nil {
		return err
	}
	if err := cfg.load(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) LoadRootDirFromFlagOrDefault() error {
	if cfg.Rootdir == "" {
		rootdir := cfg.v.GetString(RootdirKey)
		if rootdir != "" {
			return cfg.setRootdir(rootdir)
		}

		return cfg.setRootdir(DefaultRootDir())
	}

	return nil
}

func (cfg *Config) setRootdir(rootdir string) error {
	var err error
	if rootdir == "" {
		return NewErrInvalidRootDir(rootdir)
	}
	// using absolute rootdir for robustness.
	cfg.Rootdir, err = filepath.Abs(rootdir)
	if err != nil {
		return err
	}
	cfg.v.AddConfigPath(cfg.Rootdir)
	return nil
}

func (cfg *Config) validate() error {
	if err := cfg.Datastore.validate(); err != nil {
		return NewErrFailedToValidateConfig(err)
	}
	if err := cfg.API.validate(); err != nil {
		return NewErrFailedToValidateConfig(err)
	}
	if err := cfg.Net.validate(); err != nil {
		return NewErrFailedToValidateConfig(err)
	}
	if err := cfg.Log.validate(); err != nil {
		return NewErrFailedToValidateConfig(err)
	}
	return nil
}

func (cfg *Config) paramsPreprocessing() error {
	// We prefer using absolute paths, relative to the rootdir.
	if !filepath.IsAbs(cfg.v.GetString("datastore.badger.path")) {
		cfg.v.Set("datastore.badger.path", filepath.Join(cfg.Rootdir, cfg.v.GetString("datastore.badger.path")))
	}
	if !filepath.IsAbs(cfg.v.GetString("api.privkeypath")) {
		cfg.v.Set("api.privkeypath", filepath.Join(cfg.Rootdir, cfg.v.GetString("api.privkeypath")))
	}
	if !filepath.IsAbs(cfg.v.GetString("api.pubkeypath")) {
		cfg.v.Set("api.pubkeypath", filepath.Join(cfg.Rootdir, cfg.v.GetString("api.pubkeypath")))
	}

	// log.logger configuration as a string
	logloggerAsStringSlice := cfg.v.GetStringSlice("log.logger")
	if logloggerAsStringSlice != nil {
		cfg.v.Set("log.logger", strings.Join(logloggerAsStringSlice, ";"))
	}

	// Expand the passed in `~` if it wasn't expanded properly by the shell.
	// That can happen when the parameters are passed from outside of a shell.
	if err := expandHomeDir(&cfg.API.PrivKeyPath); err != nil {
		return err
	}
	if err := expandHomeDir(&cfg.API.PubKeyPath); err != nil {
		return err
	}

	var bs ByteSize
	if err := bs.Set(cfg.v.GetString("datastore.badger.valuelogfilesize")); err != nil {
		return err
	}
	cfg.Datastore.Badger.ValueLogFileSize = bs

	return nil
}

func (cfg *Config) load() error {
	if err := cfg.Log.load(); err != nil {
		return err
	}
	return nil
}

// DatastoreConfig configures datastores.
type DatastoreConfig struct {
	Store         string
	Memory        MemoryConfig
	Badger        BadgerConfig
	MaxTxnRetries int
}

// BadgerConfig configures Badger's on-disk / filesystem mode.
type BadgerConfig struct {
	Path             string
	ValueLogFileSize ByteSize
	*badgerds.Options
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
		MaxTxnRetries: 5,
	}
}

func (dbcfg DatastoreConfig) validate() error {
	switch dbcfg.Store {
	case "badger", "memory":
	default:
		return NewErrInvalidDatastoreType(dbcfg.Store)
	}
	return nil
}

// APIConfig configures the API endpoints.
type APIConfig struct {
	Address        string
	TLS            bool
	AllowedOrigins []string `mapstructure:"allowed-origins"`
	PubKeyPath     string
	PrivKeyPath    string
	Email          string
}

func defaultAPIConfig() *APIConfig {
	return &APIConfig{
		Address:        "localhost:9181",
		TLS:            false,
		AllowedOrigins: []string{},
		PubKeyPath:     "certs/server.key",
		PrivKeyPath:    "certs/server.crt",
		Email:          DefaultAPIEmail,
	}
}

func (apicfg *APIConfig) validate() error {
	if apicfg.Address == "" {
		return ErrInvalidDatabaseURL
	}

	if apicfg.Address == "localhost" || net.ParseIP(apicfg.Address) != nil { //nolint:goconst
		return ErrMissingPortNumber
	}

	if isValidDomainName(apicfg.Address) {
		return nil
	}

	host, _, err := net.SplitHostPort(apicfg.Address)
	if err != nil {
		return NewErrInvalidDatabaseURL(err)
	}
	if host == "localhost" {
		return nil
	}
	if net.ParseIP(host) == nil {
		return ErrNoPortWithDomain
	}

	return nil
}

func isValidDomainName(domain string) bool {
	asciiDomain, err := idna.Registration.ToASCII(domain)
	if err != nil {
		return false
	}
	return asciiDomain == domain
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
	P2PAddress    string
	P2PDisabled   bool
	Peers         string
	PubSubEnabled bool `mapstructure:"pubsub"`
	RelayEnabled  bool `mapstructure:"relay"`
}

func defaultNetConfig() *NetConfig {
	return &NetConfig{
		P2PAddress:    "/ip4/0.0.0.0/tcp/9171",
		P2PDisabled:   false,
		Peers:         "",
		PubSubEnabled: true,
		RelayEnabled:  false,
	}
}

func (netcfg *NetConfig) validate() error {
	_, err := ma.NewMultiaddr(netcfg.P2PAddress)
	if err != nil {
		return NewErrInvalidP2PAddress(err, netcfg.P2PAddress)
	}
	if len(netcfg.Peers) > 0 {
		peers := strings.Split(netcfg.Peers, ",")
		maddrs := make([]ma.Multiaddr, len(peers))
		for i, addr := range peers {
			addr, err := ma.NewMultiaddr(addr)
			if err != nil {
				return NewErrInvalidBootstrapPeers(err, netcfg.Peers)
			}
			maddrs[i] = addr
		}
	}
	return nil
}

// LogConfig configures output and logger.
type LoggingConfig struct {
	Level          string
	Stacktrace     bool
	Format         string
	Output         string // logging actually supports multiple output paths, but here only one is supported
	Caller         bool
	NoColor        bool
	Logger         string
	NamedOverrides map[string]*NamedLoggingConfig
}

// NamedLoggingConfig is a named logging config, used for named overrides of the default config.
type NamedLoggingConfig struct {
	Name string
	LoggingConfig
}

func defaultLogConfig() *LoggingConfig {
	return &LoggingConfig{
		Level:          logLevelInfo,
		Stacktrace:     false,
		Format:         "csv",
		Output:         "stderr",
		Caller:         false,
		NoColor:        false,
		Logger:         "",
		NamedOverrides: make(map[string]*NamedLoggingConfig),
	}
}

// validate ensures that the logging config is valid.
func (logcfg *LoggingConfig) validate() error {
	/*
		`loglevel` is either a single value, or a single value with comma-separated list of key=value pairs, for which
		the key is the name of the logger and the value is the log level, each logger name is unique, and value is valid.

			`--loglevels <default>,<loggerNname>=<value>,...`
	*/
	kvs := []map[string]string{}
	validLevel := func(level string) bool {
		for _, l := range []string{
			logLevelDebug,
			logLevelInfo,
			logLevelError,
			logLevelFatal,
		} {
			if l == level {
				return true
			}
		}
		return false
	}
	ensureUniqueKeys := func(kvs []map[string]string) error {
		keys := make(map[string]bool)
		for _, kv := range kvs {
			for k := range kv {
				if keys[k] {
					return NewErrDuplicateLoggerName(k)
				}
				keys[k] = true
			}
		}
		return nil
	}

	parts := strings.Split(logcfg.Level, ",")
	if len(parts) > 0 {
		if !validLevel(parts[0]) {
			return NewErrInvalidLogLevel(parts[0])
		}
		for _, kv := range parts[1:] {
			parsedKV, err := parseKV(kv)
			if err != nil {
				return err
			}
			// ensure each value is a valid loglevel validLevel
			if !validLevel(parsedKV[1]) {
				return NewErrInvalidLogLevel(parsedKV[1])
			}
			kvs = append(kvs, map[string]string{parsedKV[0]: parsedKV[1]})
		}
		if err := ensureUniqueKeys(kvs); err != nil {
			return err
		}
	}

	// logger: expect format like: `net,nocolor=true,level=debug;config,output=stdout,level=info`
	if len(logcfg.Logger) != 0 {
		namedconfigs := strings.Split(logcfg.Logger, ";")
		for _, c := range namedconfigs {
			parts := strings.Split(c, ",")
			if len(parts) < 2 {
				return NewErrLoggerConfig("unexpected format (expected: `module,key=value;module,key=value;...`")
			}
			if parts[0] == "" {
				return ErrLoggerNameEmpty
			}
			for _, pair := range parts[1:] {
				parsedKV, err := parseKV(pair)
				if err != nil {
					return err
				}
				if !isLowercaseAlpha(parsedKV[0]) {
					return NewErrInvalidLoggerName(parsedKV[0])
				}
				switch parsedKV[0] {
				case "format", "output", "nocolor", "stacktrace", "caller": //nolint:goconst
					// valid logger parameters
				case "level": //nolint:goconst
					// ensure each value is a valid loglevel validLevel
					if !validLevel(parsedKV[1]) {
						return NewErrInvalidLogLevel(parsedKV[1])
					}
				default:
					return NewErrUnknownLoggerParameter(parsedKV[0])
				}
			}
		}
	}

	return nil
}

func (logcfg *LoggingConfig) load() error {
	// load loglevel
	parts := strings.Split(logcfg.Level, ",")
	if len(parts) > 0 {
		logcfg.Level = parts[0]
	}
	if len(parts) > 1 {
		for _, kv := range parts[1:] {
			parsedKV := strings.Split(kv, "=")
			if len(parsedKV) != 2 {
				return NewErrInvalidLogLevel(kv)
			}
			c, err := logcfg.GetOrCreateNamedLogger(parsedKV[0])
			if err != nil {
				return NewErrCouldNotObtainLoggerConfig(err, parsedKV[0])
			}
			c.Level = parsedKV[1]
		}
	}

	// load logger
	// e.g. `net,nocolor=true,level=debug;config,output=stdout,level=info`
	// logger has higher priority over loglevel whenever both touch the same parameters
	if len(logcfg.Logger) != 0 {
		s := strings.Split(logcfg.Logger, ";")
		for _, v := range s {
			vs := strings.Split(v, ",")
			override, err := logcfg.GetOrCreateNamedLogger(vs[0])
			if err != nil {
				return NewErrCouldNotObtainLoggerConfig(err, vs[0])
			}
			override.Name = vs[0]
			for _, v := range vs[1:] {
				parsedKV := strings.Split(v, "=")
				if len(parsedKV) != 2 {
					return NewErrNotProvidedAsKV(v)
				}
				switch param := strings.ToLower(parsedKV[0]); param {
				case "level": // string
					override.Level = parsedKV[1]
				case "format": // string
					override.Format = parsedKV[1]
				case "output": // string
					override.Output = parsedKV[1]
				case "stacktrace": // bool
					if override.Stacktrace, err = strconv.ParseBool(parsedKV[1]); err != nil {
						return NewErrCouldNotParseType(err, "bool")
					}
				case "nocolor": // bool
					if override.NoColor, err = strconv.ParseBool(parsedKV[1]); err != nil {
						return NewErrCouldNotParseType(err, "bool")
					}
				case "caller": // bool
					if override.Caller, err = strconv.ParseBool(parsedKV[1]); err != nil {
						return NewErrCouldNotParseType(err, "bool")
					}
				default:
					return NewErrUnknownLoggerParameter(param)
				}
			}
		}
	}

	c, err := logcfg.toLoggerConfig()
	if err != nil {
		return err
	}
	logging.SetConfig(c)
	return nil
}

func convertLoglevel(level string) (logging.LogLevel, error) {
	switch level {
	case logLevelDebug:
		return logging.Debug, nil
	case logLevelInfo:
		return logging.Info, nil
	case logLevelError:
		return logging.Error, nil
	case logLevelFatal:
		return logging.Fatal, nil
	default:
		return logging.LogLevel(0), NewErrInvalidLogLevel(level)
	}
}

// Exports the logging config to the logging library's config.
func (logcfg LoggingConfig) toLoggerConfig() (logging.Config, error) {
	loglevel, err := convertLoglevel(logcfg.Level)
	if err != nil {
		return logging.Config{}, err
	}

	var encfmt logging.EncoderFormat
	switch logcfg.Format {
	case "json":
		encfmt = logging.JSON
	case "csv":
		encfmt = logging.CSV
	default:
		return logging.Config{}, NewErrInvalidLogFormat(logcfg.Format)
	}

	// handle logger named overrides
	overrides := make(map[string]logging.Config)
	for name, cfg := range logcfg.NamedOverrides {
		c, err := cfg.toLoggerConfig()
		if err != nil {
			return logging.Config{}, NewErrOverrideConfigConvertFailed(err, name)
		}
		overrides[name] = c
	}

	c := logging.Config{
		Level:                 logging.NewLogLevelOption(loglevel),
		EnableStackTrace:      logging.NewEnableStackTraceOption(logcfg.Stacktrace),
		DisableColor:          logging.NewDisableColorOption(logcfg.NoColor),
		EncoderFormat:         logging.NewEncoderFormatOption(encfmt),
		OutputPaths:           []string{logcfg.Output},
		EnableCaller:          logging.NewEnableCallerOption(logcfg.Caller),
		OverridesByLoggerName: overrides,
	}
	return c, nil
}

// this is a copy that doesn't deep copy the NamedOverrides map
// copy is handled by runtime "pass-by-value"
func (logcfg LoggingConfig) copy() LoggingConfig {
	logcfg.NamedOverrides = make(map[string]*NamedLoggingConfig)
	return logcfg
}

// GetOrCreateNamedLogger returns a named logger config, or creates a default one if it doesn't exist.
func (logcfg *LoggingConfig) GetOrCreateNamedLogger(name string) (*NamedLoggingConfig, error) {
	if name == "" {
		return nil, ErrLoggerNameEmpty
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

// BindFlag binds a CLI flag to a config key.
func (cfg *Config) BindFlag(key string, flag *pflag.Flag) error {
	return cfg.v.BindPFlag(key, flag)
}

// ToJSON serializes the config to a JSON byte array.
func (c *Config) ToJSON() ([]byte, error) {
	jsonbytes, err := json.Marshal(c)
	if err != nil {
		return []byte{}, NewErrConfigToJSONFailed(err)
	}
	return jsonbytes, nil
}

// String serializes the config to a JSON string.
func (c *Config) String() string {
	jsonbytes, err := c.ToJSON()
	if err != nil {
		return fmt.Sprintf("failed to convert config to string: %s", err)
	}
	return string(jsonbytes)
}

func (c *Config) toBytes() ([]byte, error) {
	var buffer bytes.Buffer
	tmpl := template.New("configTemplate")
	configTemplate, err := tmpl.Parse(defaultConfigTemplate)
	if err != nil {
		return nil, NewErrConfigTemplateFailed(err)
	}
	if err := configTemplate.Execute(&buffer, c); err != nil {
		return nil, NewErrConfigTemplateFailed(err)
	}
	return buffer.Bytes(), nil
}
