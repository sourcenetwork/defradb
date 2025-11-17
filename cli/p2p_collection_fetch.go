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
)

func MakeP2PCollectionFetchCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "fetch [versionID...]",
		Short: "Fetches specific collection versions from the network",
		Long: `Fetches specific collection versions from the network.

This command allows you to fetch collection versions across the network.
Older versions of a requested collection will also be fetched.
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
			return cliClient.FetchCollections(ctx, args...)
		},
	}

	EmbedCLIExample(ctx, cmd, "fetch single collection versions",
		`defradb client p2p collection fetch bafy123`)

	EmbedCLIExample(ctx, cmd, "fetch multiple collection versions",
		`defradb client p2p collection fetch bafy123 bafy456`)

	cmd.Flags().Duration("timeout", 0, "Timeout for fetch operations")
	return cmd
}
