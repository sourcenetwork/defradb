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
	"github.com/spf13/pflag"
)

// rootFlags is a set of persistent flags that are bound to config values.
var rootFlags = pflag.NewFlagSet("root", pflag.ContinueOnError)

func init() {
	rootFlags.String(
		"rootdir",
		"",
		"Directory for persistent data (default: $HOME/.defradb)",
	)
	rootFlags.String(
		"log-level",
		"info",
		"Log level to use. Options are debug, info, error, fatal",
	)
	rootFlags.String(
		"log-output",
		"stderr",
		"Log output path. Options are stderr or stdout.",
	)
	rootFlags.String(
		"log-format",
		"text",
		"Log format to use. Options are text or json",
	)
	rootFlags.Bool(
		"log-stacktrace",
		false,
		"Include stacktrace in error and fatal logs",
	)
	rootFlags.Bool(
		"log-source",
		false,
		"Include source location in logs",
	)
	rootFlags.String(
		"log-overrides",
		"",
		"Logger config overrides. Format <name>,<key>=<val>,...;<name>,...",
	)
	rootFlags.Bool(
		"no-log-color",
		false,
		"Disable colored log output",
	)
	rootFlags.String(
		"url",
		"127.0.0.1:9181",
		"URL of HTTP endpoint to listen on or connect to",
	)
	rootFlags.String(
		"keyring-namespace",
		"defradb",
		"Service name to use when using the system backend",
	)
	rootFlags.String(
		"keyring-backend",
		"file",
		"Keyring backend to use. Options are file or system",
	)
	rootFlags.String(
		"keyring-path",
		"keys",
		"Path to store encrypted keys when using the file backend",
	)
	rootFlags.Bool(
		"no-keyring",
		false,
		"Disable the keyring and generate ephemeral keys",
	)
}

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

	cmd.PersistentFlags().AddFlagSet(rootFlags)

	return cmd
}
