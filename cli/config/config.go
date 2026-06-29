// Copyright 2024 Democratized Data Foundation
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
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/sourcenetwork/corelog"
)

const (
	ConfigStoreBadger   = "badger"
	ConfigStoreMemory   = "memory"
	ConfigLogFormatJSON = "json"
	ConfigLogFormatCSV  = "csv"
	ConfigLogLevelInfo  = "info"
	ConfigLogLevelDebug = "debug"
	ConfigLogLevelError = "error"
	ConfigLogLevelFatal = "fatal"

	// defaultTLSCertDir is the directory, relative to the rootdir, that is
	// searched for the default TLS certificate and key files.
	defaultTLSCertDir = "certs"
	// defaultTLSCertFile and defaultTLSKeyFile are the certificate and private
	// key file names looked up under <rootdir>/certs to automatically enable
	// TLS for the HTTP API.
	defaultTLSCertFile = "server.crt"
	defaultTLSKeyFile  = "server.key"
)

// ConfigPaths are config keys that will be made relative to the rootdir
var ConfigPaths = []string{
	"datastore.path",
	"api.pubkeypath",
	"api.privkeypath",
	"keyring.path",
}

// ConfigFlags is a mapping of cli flag names to config keys to bind.
var ConfigFlags = map[string]string{
	"log-level":                  "log.level",
	"log-output":                 "log.output",
	"log-format":                 "log.format",
	"log-stacktrace":             "log.stacktrace",
	"log-source":                 "log.source",
	"log-overrides":              "log.overrides",
	"no-log-color":               "log.colordisabled",
	"url":                        "api.address",
	"audience":                   "api.audience",
	"max-txn-retries":            "datastore.maxtxnretries",
	"store":                      "datastore.store",
	"no-encryption":              "datastore.noencryption",
	"no-searchable-encryption":   "datastore.nosearchableencryption",
	"no-signing":                 "datastore.nosigning",
	"default-key-type":           "datastore.defaultkeytype",
	"valuelogfilesize":           "datastore.badger.valuelogfilesize",
	"peers":                      "net.peers",
	"p2paddr":                    "net.p2paddresses",
	"no-p2p":                     "net.p2pdisabled",
	"pubsub":                     "net.pubsubenabled",
	"relay":                      "net.relay",
	"allowed-origins":            "api.allowed-origins",
	"pubkeypath":                 "api.pubkeypath",
	"privkeypath":                "api.privkeypath",
	"keyring-namespace":          "keyring.namespace",
	"keyring-backend":            "keyring.backend",
	"keyring-path":               "keyring.path",
	"no-keyring":                 "keyring.disabled",
	"document-acp-type":          "acp.document.type",
	"source-hub-address":         "acp.document.sourceHub.address",
	"development":                "development",
	"secret-file":                "secretfile",
	"no-telemetry":               "telemetry.disabled",
	"replicator-retry-intervals": "replicator.retryintervals",
}

// ConfigDefaults contains default values for config entries.
var ConfigDefaults = map[string]any{
	"api.address":                       "127.0.0.1:9181",
	"api.audience":                      "",
	"api.allowed-origins":               []string{},
	"datastore.path":                    "data",
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
	"telemetry.disabled":                false,
	"datastore.nosigning":               false,
	"datastore.nosearchableencryption":  false,
	"datastore.defaultkeytype":          "secp256k1",
	"acp.document.type":                 "local",
	"replicator.retryintervals":         []int{30, 60, 120, 240, 480, 960, 1920},
}

// DefaultConfig returns a new config with default values.
func DefaultConfig() *viper.Viper {
	cfg := viper.New()

	cfg.AutomaticEnv()
	cfg.SetEnvPrefix("DEFRA")
	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	cfg.SetConfigName("config")
	cfg.SetConfigType("yaml")

	for key, val := range ConfigDefaults {
		cfg.SetDefault(key, val)
	}
	return cfg
}

// CreateConfig writes the default config file if one does not exist.
func CreateConfig(rootdir string, flags *pflag.FlagSet) error {
	cfg := DefaultConfig()
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

// LoadConfig returns a new config with values from the config in the given rootdir.
func LoadConfig(rootdir string, flags *pflag.FlagSet) (*viper.Viper, error) {
	cfg := DefaultConfig()
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
	for _, key := range ConfigPaths {
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

	// Automatically enable TLS when the default certificate files are present
	// and the user has not configured certificate paths explicitly. Done after
	// the logging config is applied so any warning is formatted as configured.
	autoDetectTLSCertPaths(cfg, rootdir)

	return cfg, nil
}

// autoDetectTLSCertPaths sets the TLS certificate and key paths from the
// default certs directory (<rootdir>/certs) when the user has not configured
// them explicitly. It sets whichever of server.crt (pubkeypath) and server.key
// (privkeypath) are present: when both exist TLS is enabled, and when only one
// exists the resulting incomplete pair is rejected by the start command, so a
// half-populated certs directory is a clear error rather than a silently
// ignored certificate.
func autoDetectTLSCertPaths(cfg *viper.Viper, rootdir string) {
	// Respect explicitly-configured paths (flag, config file, or env var); if
	// either is set the user is in control of TLS and we must not override it.
	if cfg.GetString("api.pubkeypath") != "" || cfg.GetString("api.privkeypath") != "" {
		return
	}

	certPath := filepath.Join(rootdir, defaultTLSCertDir, defaultTLSCertFile)
	keyPath := filepath.Join(rootdir, defaultTLSCertDir, defaultTLSKeyFile)
	if isRegularFile(certPath) {
		cfg.Set("api.pubkeypath", certPath)
	}
	if isRegularFile(keyPath) {
		cfg.Set("api.privkeypath", keyPath)
	}
}

// isRegularFile reports whether the path exists and is a regular file
// (symlinks are followed). Directories and other irregular files return false.
func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

// bindConfigFlags binds the set of cli flags to config values.
func bindConfigFlags(cfg *viper.Viper, flags *pflag.FlagSet) error {
	var errs []error
	flags.VisitAll(func(f *pflag.Flag) {
		errs = append(errs, cfg.BindPFlag(ConfigFlags[f.Name], f))
	})
	return errors.Join(errs...)
}
