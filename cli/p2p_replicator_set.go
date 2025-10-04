// Copyright 2023 Democratized Data Foundation
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
	"strings"

	"github.com/spf13/cobra"
)

func MakeP2PReplicatorSetCommand() *cobra.Command {
	var collections []string
	var cmd = &cobra.Command{
		Use:   "set [-c, --collection] <peer>",
		Short: "Add replicator(s) and start synchronization",
		Long: `Add replicator(s) and start synchronization.
A replicator synchronizes one or all collection(s) from this node to another.

Example:
  defradb client p2p replicator set -c Users /ip4/0.0.0.0/tcp/9171/p2p/12D3KooW...
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			var addresses []string
			for _, id := range strings.Split(args[0], ",") {
				id = strings.TrimSpace(id)
				if id == "" {
					continue
				}
				addresses = append(addresses, id)
			}
			return cliClient.SetReplicator(cmd.Context(), addresses, collections...)
		},
	}

	cmd.Flags().StringSliceVarP(&collections, "collection", "c",
		[]string{}, "Collection(s) to replicate")
	return cmd
}
