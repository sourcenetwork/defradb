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

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/config"
)

func MakeRootCommand(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "defradb",
		Short: "DefraDB Edge Database",
		Long: `DefraDB is the edge database to power the user-centric future.

Start a DefraDB node, interact with a local or remote node, and much more.
`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return loadConfig(cfg)
		},
	}

	cmd.PersistentFlags().String(
		"rootdir", "",
		"Directory for data and configuration to use (default: $HOME/.defradb)",
	)
	err := cfg.BindFlag(config.RootdirKey, cmd.PersistentFlags().Lookup("rootdir"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind rootdir", err)
	}

	cmd.PersistentFlags().String(
		"loglevel", cfg.Log.Level,
		"Log level to use. Options are debug, info, error, fatal",
	)
	err = cfg.BindFlag("log.level", cmd.PersistentFlags().Lookup("loglevel"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind log.loglevel", err)
	}

	cmd.PersistentFlags().StringArray(
		"logger", []string{},
		"Override logger parameters. Usage: --logger <name>,level=<level>,output=<output>,...",
	)
	err = cfg.BindFlag("log.logger", cmd.PersistentFlags().Lookup("logger"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind log.logger", err)
	}

	cmd.PersistentFlags().String(
		"logoutput", cfg.Log.Output,
		"Log output path",
	)
	err = cfg.BindFlag("log.output", cmd.PersistentFlags().Lookup("logoutput"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind log.output", err)
	}

	cmd.PersistentFlags().String(
		"logformat", cfg.Log.Format,
		"Log format to use. Options are csv, json",
	)
	err = cfg.BindFlag("log.format", cmd.PersistentFlags().Lookup("logformat"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind log.format", err)
	}

	cmd.PersistentFlags().Bool(
		"logtrace", cfg.Log.Stacktrace,
		"Include stacktrace in error and fatal logs",
	)
	err = cfg.BindFlag("log.stacktrace", cmd.PersistentFlags().Lookup("logtrace"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind log.stacktrace", err)
	}

	cmd.PersistentFlags().Bool(
		"lognocolor", cfg.Log.NoColor,
		"Disable colored log output",
	)
	err = cfg.BindFlag("log.nocolor", cmd.PersistentFlags().Lookup("lognocolor"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind log.nocolor", err)
	}

	cmd.PersistentFlags().String(
		"url", cfg.API.Address,
		"URL of HTTP endpoint to listen on or connect to",
	)
	err = cfg.BindFlag("api.address", cmd.PersistentFlags().Lookup("url"))
	if err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind api.address", err)
	}

	return cmd
}
