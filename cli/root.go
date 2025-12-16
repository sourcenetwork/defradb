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

	"github.com/sourcenetwork/defradb/cli/config"
)

func MakeRootCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		SilenceUsage: true,
		Use:          "defradb",
		Short:        "DefraDB Edge Database",
		Long: `DefraDB is the edge database to power the user-centric future.

Start a DefraDB node, interact with a local or remote node, and much more.
`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := setContextRootDir(cmd); err != nil {
				return err
			}
			return setContextConfig(cmd)
		},
	}
	// set default flag values from config
	cfg := config.DefaultConfig()
	cmd.PersistentFlags().String(
		"rootdir",
		"",
		"Directory for persistent data (default: $HOME/.defradb)",
	)
	cmd.PersistentFlags().String(
		"log-level",
		cfg.GetString(config.ConfigFlags["log-level"]),
		"Log level to use. Options are debug, info, error, fatal",
	)
	cmd.PersistentFlags().String(
		"log-output",
		cfg.GetString(config.ConfigFlags["log-output"]),
		"Log output path. Options are stderr or stdout.",
	)
	cmd.PersistentFlags().String(
		"log-format",
		cfg.GetString(config.ConfigFlags["log-format"]),
		"Log format to use. Options are text or json",
	)
	cmd.PersistentFlags().Bool(
		"log-stacktrace",
		cfg.GetBool(config.ConfigFlags["log-stacktrace"]),
		"Include stacktrace in error and fatal logs",
	)
	cmd.PersistentFlags().Bool(
		"log-source",
		cfg.GetBool(config.ConfigFlags["log-source"]),
		"Include source location in logs",
	)
	cmd.PersistentFlags().String(
		"log-overrides",
		cfg.GetString(config.ConfigFlags["log-overrides"]),
		"Logger config overrides. Format <name>,<key>=<val>,...;<name>,...",
	)
	cmd.PersistentFlags().Bool(
		"no-log-color",
		cfg.GetBool(config.ConfigFlags["no-log-color"]),
		"Disable colored log output",
	)
	cmd.PersistentFlags().String(
		"url",
		cfg.GetString(config.ConfigFlags["url"]),
		"URL of HTTP endpoint to listen on or connect to",
	)
	cmd.PersistentFlags().String(
		"keyring-namespace",
		cfg.GetString(config.ConfigFlags["keyring-namespace"]),
		"Service name to use when using the system backend",
	)
	cmd.PersistentFlags().String(
		"keyring-backend",
		cfg.GetString(config.ConfigFlags["keyring-backend"]),
		"Keyring backend to use. Options are file or system",
	)
	cmd.PersistentFlags().String(
		"keyring-path",
		cfg.GetString(config.ConfigFlags["keyring-path"]),
		"Path to store encrypted keys when using the file backend",
	)
	cmd.PersistentFlags().Bool(
		"no-keyring",
		cfg.GetBool(config.ConfigFlags["no-keyring"]),
		"Disable the keyring and generate ephemeral keys",
	)
	cmd.PersistentFlags().String(
		"source-hub-address",
		cfg.GetString(config.ConfigFlags["source-hub-address"]),
		"The SourceHub address authorized by the client to make SourceHub transactions on behalf of the actor",
	)
	cmd.PersistentFlags().String(
		"secret-file",
		cfg.GetString(config.ConfigFlags["secret-file"]),
		"Path to the file containing secrets")
	return cmd
}
