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
	"context"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client/options"
)

func MakeP2PReplicatorDeleteCommand(ctx context.Context) *cobra.Command {
	var collections []string
	var cmd = &cobra.Command{
		Use:   "delete [-c, --collection] <peerID>",
		Short: "Delete replicator(s) and stop synchronization",
		Long: `Delete replicator(s) and stop synchronization.
A replicator synchronizes one or all collection(s) from this instance to another.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)
			opt := options.WithIdentity(
				options.DeleteReplicator().SetCollectionNames(collections),
				identity.FromContext(cmd.Context()),
			)
			return cliClient.DeleteReplicator(cmd.Context(), args[0], opt)
		},
	}

	EmbedCLIExample(ctx, cmd, "Delete replicator",
		`defradb client p2p replicator delete -c Users 12D3...`)

	cmd.Flags().StringSliceVarP(&collections, "collection", "c",
		[]string{}, "Collection(s) to stop replicating")
	return cmd
}
