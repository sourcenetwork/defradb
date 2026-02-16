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

func MakeP2PReplicatorCreateCommand(ctx context.Context) *cobra.Command {
	var collections []string
	var cmd = &cobra.Command{
		Use:   "create [-c, --collection] <addresses...>",
		Short: "Create replicator(s) and start synchronization",
		Long: `Create replicator(s) and start synchronization.
A replicator synchronizes one or all collection(s) from this instance to another.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)
			opt := options.WithIdentity(
				options.CreateReplicator().SetCollectionNames(collections),
				identity.FromContext(cmd.Context()),
			)
			return cliClient.CreateReplicator(cmd.Context(), args, opt)
		},
	}

	EmbedCLIExample(ctx, cmd, "Create a replicator to replicate the \"Users\" collection to a peer",
		`defradb client p2p replicator create -c Users /ip4/0.0.0.0/tcp/9171/p2p/12D3Ko...`)

	EmbedCLIExample(
		ctx,
		cmd,
		"Create a replicator to replicate the \"Orders\" collection to multiple peers",
		`defradb client p2p replicator create -c Orders `+
			`/ip4/0.0.0.0/tcp/9171/p2p/12D3Ko... `+
			`/ip4/0.0.0.0/tcp/9172/p2p/1543LK...`,
	)

	cmd.Flags().StringSliceVarP(&collections, "collection", "c",
		[]string{}, "Collection(s) to replicate")
	return cmd
}
