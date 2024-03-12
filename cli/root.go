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

	cmd.PersistentFlags().String(
		"url",
		"127.0.0.1:9181",
		"URL of HTTP endpoint to listen on or connect to",
	)

	cmd.PersistentFlags().StringArray(
		"peers",
		[]string{},
		"List of peers to connect to",
	)

	cmd.PersistentFlags().Int(
		"max-txn-retries",
		5,
		"Specify the maximum number of retries per transaction",
	)

	cmd.PersistentFlags().String(
		"store",
		"badger",
		"Specify the datastore to use (supported: badger, memory)",
	)

	cmd.PersistentFlags().Int(
		"valuelogfilesize",
		1<<30,
		"Specify the datastore value log file size (in bytes). In memory size will be 2*valuelogfilesize",
	)

	cmd.PersistentFlags().StringSlice(
		"p2paddr",
		[]string{"/ip4/127.0.0.1/tcp/9171"},
		"Listen addresses for the p2p network (formatted as a libp2p MultiAddr)",
	)

	cmd.PersistentFlags().Bool(
		"no-p2p",
		false,
		"Disable the peer-to-peer network synchronization system",
	)

	cmd.PersistentFlags().StringArray(
		"allowed-origins",
		[]string{},
		"List of origins to allow for CORS requests",
	)

	cmd.PersistentFlags().String(
		"pubkeypath",
		"",
		"Path to the public key for tls",
	)

	cmd.PersistentFlags().String(
		"privkeypath",
		"",
		"Path to the private key for tls",
	)

	return cmd
}
