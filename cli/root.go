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
	"github.com/spf13/cobra"
)

func MakeRootCommand() *cobra.Command {
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

	cmd.PersistentFlags().String(
		"rootdir",
		"",
		"Directory for persistent data (default: $HOME/.defradb)",
	)

	cmd.PersistentFlags().String(
		"log-level",
		"info",
		"Log level to use. Options are debug, info, error, fatal",
	)

	cmd.PersistentFlags().String(
		"log-output",
		"stderr",
		"Log output path. Options are stderr or stdout.",
	)

	cmd.PersistentFlags().String(
		"log-format",
		"text",
		"Log format to use. Options are text or json",
	)

	cmd.PersistentFlags().Bool(
		"log-stacktrace",
		false,
		"Include stacktrace in error and fatal logs",
	)

	cmd.PersistentFlags().Bool(
		"log-source",
		false,
		"Include source location in logs",
	)

	cmd.PersistentFlags().String(
		"log-overrides",
		"",
		"Logger config overrides. Format <name>,<key>=<val>,...;<name>,...",
	)

	cmd.PersistentFlags().Bool(
		"log-no-color",
		false,
		"Disable colored log output",
	)

	cmd.PersistentFlags().String(
		"url",
		"127.0.0.1:9181",
		"URL of HTTP endpoint to listen on or connect to",
	)

	cmd.PersistentFlags().String(
		"keyring-namespace",
		"defradb",
		"Service name to use when using the system backend",
	)

	cmd.PersistentFlags().String(
		"keyring-backend",
		"file",
		"Keyring backend to use. Options are file or system",
	)

	cmd.PersistentFlags().String(
		"keyring-path",
		"keys",
		"Path to store encrypted keys when using the file backend",
	)

	cmd.PersistentFlags().Bool(
		"no-keyring",
		false,
		"Disable the keyring and generate ephemeral keys",
	)

	return cmd
}
