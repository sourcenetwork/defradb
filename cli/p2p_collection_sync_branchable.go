// Copyright 2025 Democratized Data Foundation
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

func MakeP2PCollectionSyncBranchableCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sync-branchable [collection-id]",
		Short: "Synchronize a branchable collection's DAG from the network",
		Long: `Synchronize a branchable collection's DAG from the network.

This command allows you to sync the collection-level history for branchable collections
(collections marked with @branchable directive). It doesn't automatically subscribe
to the collection for future updates.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			collectionID := args[0]

			ctx := cmd.Context()
			if timeout, _ := cmd.Flags().GetDuration("timeout"); timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}

			cliClient := mustGetContextCLIClient(cmd)
			opt := options.WithIdentity(options.SyncBranchableCollection(), identity.FromContext(cmd.Context()))
			return cliClient.SyncBranchableCollection(ctx, collectionID, opt)
		},
	}

	EmbedCLIExample(ctx, cmd, "sync branchable collection",
		`defradb client p2p collection sync-branchable bafkreig27seqzxvr7isblvj77wvqnmkzoyv3u4nwytyethkbcpxlrx3iqq`)

	EmbedCLIExample(ctx, cmd, "sync branchable collection with timeout",
		`defradb client p2p collection sync-branchable bafkreig27seqzxvr7isblvj77wvqnmkzoyv3u4nwytyethkbcpxlrx3iqq `+
			`--timeout 10s`)

	cmd.Flags().Duration("timeout", 0, "Timeout for sync operations (default: 5s if not specified)")
	return cmd
}
