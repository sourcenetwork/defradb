// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cmd

import (
	"context"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/spf13/cobra"
)

var (
	log     = logging.MustNewLogger("defra.cli")
	cfg     = config.DefaultConfig()
	rootDir string
)

var RootCmd = rootCmd

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(context.Background(), "execution of root command failed", logging.NewKV("error", err))
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
		rootDir, exists, err := config.GetRootDir(rootDir)
		if err != nil {
			log.Fatal(ctx, "could not get rootdir", logging.NewKV("error", err))
		}
		if exists {
			err := cfg.Load(rootDir)
			if err != nil {
				log.FatalE(ctx, "could not load config file", err)
			}
			loggingConfig, err := cfg.GetLoggingConfig()
			if err != nil {
				log.FatalE(ctx, "could not get logging config", err)
			}
			logging.SetConfig(loggingConfig)
			log.Debug(
				ctx,
				"Configuration loaded from DefraDB root directory",
				logging.NewKV("rootdir", rootDir),
			)
		} else {
			err := cfg.LoadWithoutRootDir()
			if err != nil {
				log.FatalE(ctx, "could not load config file", err)
			}
			loggingConfig, err := cfg.GetLoggingConfig()
			if err != nil {
				log.FatalE(ctx, "could not get logging config", err)
			}
			logging.SetConfig(loggingConfig)
			log.Info(
				ctx,
				"Using default configuration. To create DefraDB's root directory, use defradb init.",
			)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&rootDir,
		"rootdir",
		"",
		"DefraDB's root directory (default \"$HOME/.defradb\")",
	)
	rootCmd.PersistentFlags().StringVar(
		&cfg.Logging.Level,
		"loglevel",
		cfg.Logging.Level,
		"Log level to use. Options are debug, info, warn, error, fatal",
	)
	rootCmd.PersistentFlags().StringVar(
		&cfg.Logging.OutputPath,
		"logoutput",
		cfg.Logging.OutputPath,
		"Log output path",
	)
	rootCmd.PersistentFlags().StringVar(
		&cfg.Logging.Format,
		"logformat",
		cfg.Logging.Format,
		"Log format",
	)
	rootCmd.PersistentFlags().BoolVar(
		&cfg.Logging.Stacktrace,
		"stacktrace",
		cfg.Logging.Stacktrace,
		"Include stacktrace in error and fatal logs",
	)
	rootCmd.PersistentFlags().StringVar(
		&cfg.API.Address,
		"url",
		cfg.API.Address,
		"URL of the target database's HTTP endpoint",
	)
	rootCmd.PersistentFlags().BoolVar(
		&cfg.Logging.Color,
		"color",
		cfg.Logging.Color,
		"Toggle colored output",
	)
}
