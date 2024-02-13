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
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/sourcenetwork/defradb/logging"
)

const (
	configStoreBadger = "badger"
	configStoreMemory = "memory"
)

// configPaths are config keys that will be made relative to the rootdir
var configPaths = []string{
	"datastore.badger.path",
	"api.pubkeypath",
	"api.privkeypath",
}

// configFlags is a mapping of config keys to cli flags to bind to.
var configFlags = map[string]string{
	"log.level":                         "loglevel",
	"log.output":                        "logoutput",
	"log.format":                        "logformat",
	"log.stacktrace":                    "logtrace",
	"log.nocolor":                       "lognocolor",
	"api.address":                       "url",
	"datastore.maxtxnretries":           "max-txn-retries",
	"datastore.store":                   "store",
	"datastore.badger.valuelogfilesize": "valuelogfilesize",
	"net.peers":                         "peers",
	"net.p2paddresses":                  "p2paddr",
	"net.p2pdisabled":                   "no-p2p",
	"api.allowed-origins":               "allowed-origins",
	"api.pubkeypath":                    "pubkeypath",
	"api.privkeypath":                   "privkeypath",
}

// defaultConfig returns a new config with default values.
func defaultConfig() *viper.Viper {
	cfg := viper.New()

	cfg.AutomaticEnv()
	cfg.SetEnvPrefix("DEFRA")
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	cfg.SetConfigName("config")
	cfg.SetConfigType("yaml")

	cfg.SetDefault("datastore.badger.path", "data")
	cfg.SetDefault("net.pubSubEnabled", true)
	cfg.SetDefault("net.relay", false)
	cfg.SetDefault("log.caller", false)

	return cfg
}

// createConfig writes the default config file if one does not exist.
func createConfig(rootdir string, flags *pflag.FlagSet) error {
	cfg := defaultConfig()
	cfg.AddConfigPath(rootdir)

	if err := bindConfigFlags(cfg, flags); err != nil {
		return err
	}
	// make sure rootdir exists
	if err := os.MkdirAll(rootdir, 0755); err != nil {
		return err
	}
	err := cfg.SafeWriteConfig()
	if _, ok := err.(viper.ConfigFileAlreadyExistsError); ok { //nolint:errorlint
		return nil
	}
	return err
}

// loadConfig returns a new config with values from the config in the given rootdir.
func loadConfig(rootdir string, flags *pflag.FlagSet) (*viper.Viper, error) {
	cfg := defaultConfig()
	cfg.AddConfigPath(rootdir)

	// attempt to read the existing config
	err := cfg.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok { //nolint:errorlint
		return nil, err
	}
	// bind cli flags to config keys
	if err := bindConfigFlags(cfg, flags); err != nil {
		return nil, err
	}

	// make paths relative to the rootdir
	for _, key := range configPaths {
		path := cfg.GetString(key)
		if path != "" && !filepath.IsAbs(path) {
			cfg.Set(key, filepath.Join(rootdir, path))
		}
	}

	logCfg := loggingConfig(cfg.Sub("log"))
	logCfg.OverridesByLoggerName = make(map[string]logging.Config)

	// apply named logging overrides
	for key := range cfg.GetStringMap("log.overrides") {
		logCfg.OverridesByLoggerName[key] = loggingConfig(cfg.Sub("log.overrides." + key))
	}
	logging.SetConfig(logCfg)

	return cfg, nil
}

// bindConfigFlags binds the set of cli flags to config values.
func bindConfigFlags(cfg *viper.Viper, flags *pflag.FlagSet) error {
	for key, flag := range configFlags {
		err := cfg.BindPFlag(key, flags.Lookup(flag))
		if err != nil {
			return err
		}
	}
	return nil
}

// loggingConfig returns a new logging config from the given config.
func loggingConfig(cfg *viper.Viper) logging.Config {
	var level int8
	switch value := cfg.GetString("level"); value {
	case "debug":
		level = logging.Debug
	case "info":
		level = logging.Info
	case "error":
		level = logging.Error
	case "fatal":
		level = logging.Fatal
	default:
		level = logging.Info
	}

	var format logging.EncoderFormat
	switch value := cfg.GetString("format"); value {
	case "json": //nolint:goconst
		format = logging.JSON
	case "csv":
		format = logging.CSV
	default:
		format = logging.CSV
	}

	return logging.Config{
		Level:            logging.NewLogLevelOption(level),
		EnableStackTrace: logging.NewEnableStackTraceOption(cfg.GetBool("stacktrace")),
		DisableColor:     logging.NewDisableColorOption(cfg.GetBool("nocolor")),
		EncoderFormat:    logging.NewEncoderFormatOption(format),
		OutputPaths:      []string{cfg.GetString("output")},
		EnableCaller:     logging.NewEnableCallerOption(cfg.GetBool("caller")),
	}
}
