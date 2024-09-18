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
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sourcenetwork/corelog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	configStoreBadger   = "badger"
	configStoreMemory   = "memory"
	configLogFormatJSON = "json"
	configLogFormatCSV  = "csv"
	configLogLevelInfo  = "info"
	configLogLevelDebug = "debug"
	configLogLevelError = "error"
	configLogLevelFatal = "fatal"
)

// configPaths are config keys that will be made relative to the rootdir
var configPaths = []string{
	"datastore.badger.path",
	"api.pubkeypath",
	"api.privkeypath",
	"keyring.path",
}

// configFlags is a mapping of cli flag names to config keys to bind.
var configFlags = map[string]string{
	"log-level":          "log.level",
	"log-output":         "log.output",
	"log-format":         "log.format",
	"log-stacktrace":     "log.stacktrace",
	"log-source":         "log.source",
	"log-overrides":      "log.overrides",
	"no-log-color":       "log.colordisabled",
	"url":                "api.address",
	"max-txn-retries":    "datastore.maxtxnretries",
	"store":              "datastore.store",
	"no-encryption":      "datastore.noencryption",
	"valuelogfilesize":   "datastore.badger.valuelogfilesize",
	"peers":              "net.peers",
	"p2paddr":            "net.p2paddresses",
	"no-p2p":             "net.p2pdisabled",
	"allowed-origins":    "api.allowed-origins",
	"pubkeypath":         "api.pubkeypath",
	"privkeypath":        "api.privkeypath",
	"keyring-namespace":  "keyring.namespace",
	"keyring-backend":    "keyring.backend",
	"keyring-path":       "keyring.path",
	"no-keyring":         "keyring.disabled",
	"source-hub-address": "acp.sourceHub.address",
	"development":        "development",
	"secret-file":        "secretfile",
}

// configDefaults contains default values for config entries.
var configDefaults = map[string]any{
	"api.address":                       "127.0.0.1:9181",
	"api.allowed-origins":               []string{},
	"datastore.badger.path":             "data",
	"datastore.maxtxnretries":           5,
	"datastore.store":                   "badger",
	"datastore.badger.valuelogfilesize": 1 << 30,
	"development":                       false,
	"net.p2pdisabled":                   false,
	"net.p2paddresses":                  []string{"/ip4/127.0.0.1/tcp/9171"},
	"net.peers":                         []string{},
	"net.pubSubEnabled":                 true,
	"net.relay":                         false,
	"keyring.backend":                   "file",
	"keyring.disabled":                  false,
	"keyring.namespace":                 "defradb",
	"keyring.path":                      "keys",
	"log.caller":                        false,
	"log.colordisabled":                 false,
	"log.format":                        "text",
	"log.level":                         "info",
	"log.output":                        "stderr",
	"log.source":                        false,
	"log.stacktrace":                    false,
	"secretfile":                        ".env",
}

// defaultConfig returns a new config with default values.
func defaultConfig() *viper.Viper {
	cfg := viper.New()

	cfg.AutomaticEnv()
	cfg.SetEnvPrefix("DEFRA")
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	cfg.SetConfigName("config")
	cfg.SetConfigType("yaml")

	for key, val := range configDefaults {
		cfg.SetDefault(key, val)
	}
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
	// error type is known and shouldn't be wrapped
	//
	//nolint:errorlint
	if _, ok := err.(viper.ConfigFileAlreadyExistsError); ok {
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
	// error type is known and shouldn't be wrapped
	//
	//nolint:errorlint
	if _, ok := err.(viper.ConfigFileNotFoundError); err != nil && !ok {
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

	// load environment variables from .env file if one exists
	err = godotenv.Load(cfg.GetString("secretfile"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	// set logging config
	corelog.SetConfig(corelog.Config{
		Level:            cfg.GetString("log.level"),
		Format:           cfg.GetString("log.format"),
		Output:           cfg.GetString("log.output"),
		EnableStackTrace: cfg.GetBool("log.stacktrace"),
		EnableSource:     cfg.GetBool("log.source"),
		DisableColor:     cfg.GetBool("log.colordisabled"),
	})

	// set logging config overrides
	corelog.SetConfigOverrides(cfg.GetString("log.overrides"))

	return cfg, nil
}

// bindConfigFlags binds the set of cli flags to config values.
func bindConfigFlags(cfg *viper.Viper, flags *pflag.FlagSet) error {
	var errs []error
	flags.VisitAll(func(f *pflag.Flag) {
		errs = append(errs, cfg.BindPFlag(configFlags[f.Name], f))
	})
	return errors.Join(errs...)
}
