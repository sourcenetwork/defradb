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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
)

var rootDirParam string

var rootCmd = &cobra.Command{
	Use:   "defradb",
	Short: "DefraDB Edge Database",
	Long: `DefraDB is the edge database to power the user-centric future.

Start a database node, issue a request to a local or remote node, and much more.

DefraDB is released under the BSL license, (c) 2022 Democratized Data Foundation.
See https://docs.source.network/BSL.txt for more information.
`,
	// Runs on subcommands before their Run function, to handle configuration and top-level flags.
	// Loads the rootDir containing the configuration file, otherwise warn about it and load a default configuration.
	// This allows some subcommands (`init`, `start`) to override the PreRun to create a rootDir by default.
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		rootDir, exists, err := config.GetRootDir(rootDirParam)
		if err != nil {
			return errors.Wrap("failed to get root dir", err)
		}
		defaultConfig := false
		if exists {
			err := cfg.Load(rootDir)
			if err != nil {
				return errors.Wrap("failed to load config", err)
			}
		} else {
			err := cfg.LoadWithoutRootDir()
			if err != nil {
				return errors.Wrap("failed to load config", err)
			}
			defaultConfig = true
		}

		// parse loglevel overrides
		// we use `cfg.Logging.Level` as an argument since the viper.Bind already handles
		// binding the flags / EnvVars to the struct
		if err := parseAndConfigLog(cmd.Context(), cfg.Log, cmd); err != nil {
			return err
		}

		if defaultConfig {
			log.FeedbackInfo(cmd.Context(), "Using default configuration")
		} else {
			log.FeedbackInfo(cmd.Context(), fmt.Sprintf("Configuration loaded from DefraDB directory %v", rootDir))
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&rootDirParam, "rootdir", "",
		"Directory for data and configuration to use (default \"$HOME/.defradb\")",
	)

	rootCmd.PersistentFlags().String(
		"loglevel", cfg.Log.Level,
		"Log level to use. Options are debug, info, error, fatal",
	)
	err := viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("loglevel"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind logging.loglevel", err)
	}

	rootCmd.PersistentFlags().StringArray(
		"logger", []string{},
		"Override logger parameters. Usage: --logger <name>,level=<level>,output=<output>,...",
	)

	rootCmd.PersistentFlags().String(
		"logoutput", cfg.Log.Output,
		"Log output path",
	)
	err = viper.BindPFlag("log.output", rootCmd.PersistentFlags().Lookup("logoutput"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind log.output", err)
	}

	rootCmd.PersistentFlags().String(
		"logformat", cfg.Log.Format,
		"Log format to use. Options are csv, json",
	)
	err = viper.BindPFlag("log.format", rootCmd.PersistentFlags().Lookup("logformat"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind log.format", err)
	}

	rootCmd.PersistentFlags().Bool(
		"logtrace", cfg.Log.Stacktrace,
		"Include stacktrace in error and fatal logs",
	)
	err = viper.BindPFlag("log.stacktrace", rootCmd.PersistentFlags().Lookup("logtrace"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind log.stacktrace", err)
	}

	rootCmd.PersistentFlags().Bool(
		"lognocolor", cfg.Log.NoColor,
		"Disable colored log output",
	)
	err = viper.BindPFlag("log.nocolor", rootCmd.PersistentFlags().Lookup("lognocolor"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind log.nocolor", err)
	}

	rootCmd.PersistentFlags().String(
		"url", cfg.API.Address,
		"URL of HTTP endpoint to listen on or connect to",
	)
	err = viper.BindPFlag("api.address", rootCmd.PersistentFlags().Lookup("url"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind api.address", err)
	}
}
