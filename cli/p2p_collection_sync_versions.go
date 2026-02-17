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

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client/options"
)

func MakeP2PCollectionSyncVersionsCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sync-versions [versionID...]",
		Short: "Synchronize specific collection versions from the network",
		Long: `Synchronize specific collection versions from the network.

This command allows you to synchronize collection versions across the network.
Older versions of a requested collection will also be synchronized.
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if timeout, _ := cmd.Flags().GetDuration("timeout"); timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}

			cliClient := mustGetContextCLIClient(cmd)
			opt := options.WithIdentity(options.SyncCollectionVersions(), identity.FromContext(cmd.Context()))
			return cliClient.SyncCollectionVersions(ctx, args, opt)
		},
	}

	EmbedCLIExample(ctx, cmd, "synchronize single collection versions",
		`defradb client p2p collection sync-versions bafy123`)

	EmbedCLIExample(ctx, cmd, "synchronize multiple collection versions",
		`defradb client p2p collection sync-versions bafy123 bafy456`)

	cmd.Flags().Duration("timeout", 0, "Timeout for sync operations")
	return cmd
}
