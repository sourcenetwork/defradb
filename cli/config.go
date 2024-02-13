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
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sourcenetwork/defradb/logging"
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
	"datastore.badger.inMemory":         "in-memory",
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

	cfg.SetDefault("datastore.maxTxnRetries", 5)

	cfg.SetDefault("datastore.badger.path", "data")
	cfg.SetDefault("datastore.badger.inMemory", false)
	cfg.SetDefault("datastore.badger.valueLogFileSize", 1<<30)

	cfg.SetDefault("api.address", "127.0.0.1:9181")
	cfg.SetDefault("api.pubKeyPath", "")
	cfg.SetDefault("api.privKeyPath", "")
	cfg.SetDefault("api.allowed-origins", []string{})

	cfg.SetDefault("net.p2pDisabled", false)
	cfg.SetDefault("net.p2pAddresses", []any{"/ip4/127.0.0.1/tcp/9171"})
	cfg.SetDefault("net.pubSubEnabled", true)
	cfg.SetDefault("net.relay", false)
	cfg.SetDefault("net.peers", []string{})

	cfg.SetDefault("log.level", "info")
	cfg.SetDefault("log.stackTrace", true)
	cfg.SetDefault("log.format", "csv")
	cfg.SetDefault("log.output", "stderr")
	cfg.SetDefault("log.noColor", false)
	cfg.SetDefault("log.caller", false)

	return cfg
}

// loadConfig returns a new config with values from the config in the given rootdir.
func loadConfig(rootdir string, cmd *cobra.Command, create bool) (*viper.Viper, error) {
	// rootdir defaults to $HOME/.defradb
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	if rootdir == "" {
		rootdir = filepath.Join(home, ".defradb")
	}

	cfg := defaultConfig()
	cfg.AddConfigPath(rootdir)

	// make sure rootdir exists
	if err := os.MkdirAll(rootdir, 0755); err != nil {
		return nil, err
	}
	// attempt to read the existing config
	err = cfg.ReadInConfig()
	if create && errors.As(err, &viper.ConfigFileNotFoundError{}) {
		err = cfg.SafeWriteConfig()
	}
	if err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
		return nil, err
	}

	// bind cli flags to config keys
	for key, flag := range configFlags {
		err := cfg.BindPFlag(key, cmd.Root().PersistentFlags().Lookup(flag))
		if err != nil {
			return nil, err
		}
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
