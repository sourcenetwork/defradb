// Copyright 2022 Democratized Data Foundation
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
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const badgerDatastoreName = "badger"

var (
	log          = logging.MustNewLogger("defra.cli")
	cfg          = config.DefaultConfig()
	rootDirParam string
)

var RootCmd = rootCmd

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.FatalE(context.Background(), "Execution of root command failed", err)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "defradb",
	Short: "DefraDB Edge Database",
	Long: `DefraDB is the edge database to power the user-centric future.
This CLI is the main reference implementation of DefraDB. Use it to start
a new database process, query a local or remote instance, and much more.

For example:

# Start a new database instance
> defradb start `,
	// Runs on subcommands before their Run function, to handle configuration and top-level flags.
	// Loads the rootDir containing the configuration file, otherwise warn about it and load a default configuration.
	// This allows some subcommands (`init`, `start`) to override the PreRun to create a rootDir by default.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		rootDir, exists, err := config.GetRootDir(rootDirParam)
		if err != nil {
			log.FatalE(ctx, "Could not get rootdir", err)
		}
		defaultConfig := false
		if exists {
			err := cfg.Load(rootDir)
			if err != nil {
				log.FatalE(ctx, "Could not load config file", err)
			}
		} else {
			err := cfg.LoadWithoutRootDir()
			if err != nil {
				log.FatalE(ctx, "Could not load config file", err)
			}
			defaultConfig = true
		}

		// parse loglevel overrides
		// we use `cfg.Logging.Level` as an argument since the viper.Bind already handles
		// binding the flags / EnvVars to the struct
		parseAndConfigLog(ctx, cfg.Logging, cmd)

		if defaultConfig {
			log.Info(ctx, "Using default configuration")
		} else {
			log.Debug(ctx, fmt.Sprintf("Configuration loaded from DefraDB directory %v", rootDir))

		}
	},
}

func init() {
	var err error

	rootCmd.PersistentFlags().StringVar(
		&rootDirParam, "rootdir", "",
		"directory for data and configuration to use (default \"$HOME/.defradb\")",
	)

	rootCmd.PersistentFlags().String(
		"loglevel", cfg.Logging.Level,
		"log level to use. Options are debug, info, error, fatal",
	)
	err = viper.BindPFlag("logging.level", rootCmd.PersistentFlags().Lookup("loglevel"))
	if err != nil {
		log.FatalE(context.Background(), "Could not bind logging.loglevel", err)
	}

	rootCmd.PersistentFlags().String(
		"logger", "",
		"named logger parameter override. usage: --logger <name>,level=<level>,output=<output>,etc...",
	)

	rootCmd.PersistentFlags().String(
		"logoutput", cfg.Logging.OutputPath,
		"log output path",
	)
	err = viper.BindPFlag("logging.output", rootCmd.PersistentFlags().Lookup("logoutput"))
	if err != nil {
		log.FatalE(context.Background(), "Could not bind logging.output", err)
	}

	rootCmd.PersistentFlags().String(
		"logformat", cfg.Logging.Format,
		"log format to use. Options are text, json",
	)
	err = viper.BindPFlag("logging.format", rootCmd.PersistentFlags().Lookup("logformat"))
	if err != nil {
		log.FatalE(context.Background(), "Could not bind logging.format", err)
	}

	rootCmd.PersistentFlags().Bool(
		"logtrace", cfg.Logging.Stacktrace,
		"include stacktrace in error and fatal logs",
	)
	err = viper.BindPFlag("logging.stacktrace", rootCmd.PersistentFlags().Lookup("logtrace"))
	if err != nil {
		log.FatalE(context.Background(), "Could not bind logging.stacktrace", err)
	}

	rootCmd.PersistentFlags().Bool(
		"logcolor", cfg.Logging.Color,
		"enable colored output",
	)
	err = viper.BindPFlag("logging.color", rootCmd.PersistentFlags().Lookup("logcolor"))
	if err != nil {
		log.FatalE(context.Background(), "Could not bind logging.color", err)
	}

	rootCmd.PersistentFlags().String(
		"url", cfg.API.Address,
		"URL of the target database's HTTP endpoint",
	)
	err = viper.BindPFlag("api.address", rootCmd.PersistentFlags().Lookup("url"))
	if err != nil {
		log.FatalE(context.Background(), "Could not bind api.address", err)
	}
}

// parses and then configures the given config.Config logging subconfig.
// we use log.Fatal instead of returning an error because we can't gurantee
// atomic updates, its either everything is properly set, or we Fatal()
func parseAndConfigLog(ctx context.Context, cfg *config.LoggingConfig, cmd *cobra.Command) {
	// apply logger configuration at the end
	// once everything has been processed.
	defer func() {
		loggingConfig, err := cfg.ToLoggerConfig()
		if err != nil {
			log.FatalE(ctx, "Could not get logging config", err)
		}
		logging.SetConfig(loggingConfig)
	}()

	// handle --loglevels <default>,<name>=<value>,...
	parseAndConfigLogStringParam(ctx, cfg, cfg.Level, func(l *config.LoggingConfig, v string) {
		l.Level = v
	})

	// handle --logger <name>,<field>=<value>,...
	loggerKVs, err := cmd.Flags().GetString("logger")
	if err != nil {
		log.FatalE(ctx, "can't get logger flag", err)
	}

	if loggerKVs != "" {
		parseAndConfigLogAllParams(ctx, cfg, loggerKVs)
	}
}

func parseAndConfigLogAllParams(ctx context.Context, cfg *config.LoggingConfig, kvs string) {
	if kvs == "" {
		return //nothing todo
	}

	// check if a CSV is provided
	parsed := strings.Split(kvs, ",")
	if len(parsed) <= 1 {
		log.Fatal(ctx, "invalid --logger format, must be a csv")
	}
	name := parsed[0]

	// verify KV format (<default>,<field>=<value>,...)
	// skip the first as that will be set above
	for _, kv := range parsed[1:] {
		parsedKV := strings.Split(kv, "=")
		if len(parsedKV) != 2 {
			log.Fatal(ctx, "level was not provided as <key>=<value> pair", logging.NewKV("pair", kv))
		}

		logcfg, err := cfg.GetOrCreateNamedLogger(name)
		if err != nil {
			log.FatalE(ctx, "could not get named logger config", err)
		}

		// handle field
		switch strings.ToLower(parsedKV[0]) {
		case "level": // string
			logcfg.Level = parsedKV[1]
		case "format": // string
			logcfg.Format = parsedKV[1]
		case "output": // string
			logcfg.OutputPath = parsedKV[1]
		case "stacktrace": // bool
			boolValue, err := strconv.ParseBool(parsedKV[1])
			if err != nil {
				log.FatalE(ctx, "couldn't parse kv bool", err)
			}
			logcfg.Stacktrace = boolValue
		case "color": // bool
			boolValue, err := strconv.ParseBool(parsedKV[1])
			if err != nil {
				log.FatalE(ctx, "couldn't parse kv bool", err)
			}
			logcfg.Color = boolValue
		}
	}
}

func parseAndConfigLogStringParam(ctx context.Context, cfg *config.LoggingConfig, kvs string, paramSetterFn logParamSetterStringFn) {
	if kvs == "" {
		return //nothing todo
	}

	// check if a CSV is provided
	// if its not a CSV, then just do the regular binding to the config
	parsed := strings.Split(kvs, ",")
	paramSetterFn(cfg, parsed[0])
	if len(parsed) == 1 {
		return //nothing more todo
	}

	// verify KV format (<default>,<name>=<value>,...)
	// skip the first as that will be set above
	for _, kv := range parsed[1:] {
		parsedKV := strings.Split(kv, "=")
		if len(parsedKV) != 2 {
			log.Fatal(ctx, "level was not provided as <key>=<value> pair", logging.NewKV("pair", kv))
		}

		logcfg, err := cfg.GetOrCreateNamedLogger(parsedKV[0])
		if err != nil {
			log.FatalE(ctx, "could not get named logger config", err)
		}

		paramSetterFn(&logcfg.LoggingConfig, parsedKV[1])
	}
}

//
// LEAVE FOR NOW - IMPLEMENTING SOON - PLEASE IGNORE FOR NOW
//
// func parseAndConfigLogBoolParam(ctx context.Context, cfg *config.LoggingConfig, kvs string, paramFn logParamSetterBoolFn) {
// 	if kvs == "" {
// 		return //nothing todo
// 	}

// 	// check if a CSV is provided
// 	// if its not a CSV, then just do the regular binding to the config
// 	parsed := strings.Split(kvs, ",")
// 	boolValue, err := strconv.ParseBool(parsed[0])
// 	if err != nil {
// 		log.FatalE(ctx, "couldn't parse kv bool", err)
// 	}
// 	paramFn(cfg, boolValue)
// 	if len(parsed) == 1 {
// 		return //nothing more todo
// 	}

// 	// verify KV format (<default>,<name>=<level>,...)
// 	// skip the first as that will be set above
// 	for _, kv := range parsed[1:] {
// 		parsedKV := strings.Split(kv, "=")
// 		if len(parsedKV) != 2 {
// 			log.Fatal(ctx, "field was not provided as <key>=<value> pair", logging.NewKV("pair", kv))
// 		}

// 		logcfg, err := cfg.GetOrCreateNamedLogger(parsedKV[0])
// 		if err != nil {
// 			log.FatalE(ctx, "could not get named logger config", err)
// 		}

// 		boolValue, err := strconv.ParseBool(parsedKV[1])
// 		if err != nil {
// 			log.FatalE(ctx, "couldn't parse kv bool", err)
// 		}
// 		paramFn(&logcfg.LoggingConfig, boolValue)
// 	}
// }

type logParamSetterStringFn func(*config.LoggingConfig, string)

// type logParamSetterBoolFn func(*config.LoggingConfig, bool)
