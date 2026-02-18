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

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeP2PReplicatorAddCommand(ctx context.Context) *cobra.Command {
	var collections []string
	var cmd = &cobra.Command{
		Use:   "add [-c, --collection] <addresses...>",
		Short: "Add replicator(s) and start synchronization",
		Long: `Add replicator(s) and start synchronization.
A replicator synchronizes one or all collection(s) from this instance to another.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)
			opt := options.WithIdentity(
				options.AddReplicator().SetCollectionNames(collections),
				identity.FromContext(cmd.Context()),
			)
			return cliClient.AddReplicator(cmd.Context(), args, opt)
		},
	}

	EmbedCLIExample(ctx, cmd, "Add a replicator to replicate the \"Users\" collection to a peer",
		`defradb client p2p replicator add -c Users /ip4/0.0.0.0/tcp/9171/p2p/12D3Ko...`)

	EmbedCLIExample(
		ctx,
		cmd,
		"Add a replicator to replicate the \"Orders\" collection to multiple peers",
		`defradb client p2p replicator add -c Orders `+
			`/ip4/0.0.0.0/tcp/9171/p2p/12D3Ko... `+
			`/ip4/0.0.0.0/tcp/9172/p2p/1543LK...`,
	)

	cmd.Flags().StringSliceVarP(&collections, "collection", "c",
		[]string{}, "Collection(s) to replicate")
	return cmd
}
