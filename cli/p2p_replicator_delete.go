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
	"github.com/spf13/cobra"
)

func MakeP2PReplicatorDeleteCommand() *cobra.Command {
	var collections []string
	var cmd = &cobra.Command{
		Use:   "delete [-c, --collection] <peerID>",
		Short: "Delete replicator(s) and stop synchronization",
		Long: `Delete replicator(s) and stop synchronization.
A replicator synchronizes one or all collection(s) from this instance to another.
		
Example:		
  defradb client p2p replicator delete -c Users 12D3...
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)
			return cliClient.DeleteReplicator(cmd.Context(), args[0], collections...)
		},
	}
	cmd.Flags().StringSliceVarP(&collections, "collection", "c",
		[]string{}, "Collection(s) to stop replicating")
	return cmd
}
