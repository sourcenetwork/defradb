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

func MakeP2PDocumentSyncCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sync [collection-id] [docID...]",
		Short: "Synchronize specific documents from the network",
		Long: `Synchronize specific documents from the network.

This command allows you to sync documents from a specific collection across the network.
It doesn't automatically subscribe to the collection or the documents.

Example: sync single document
  defradb client p2p document sync baf111 bae123,bae456
  `,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			collectionID := args[0]
			docIDs := args[1:]

			ctx := cmd.Context()
			if timeout, _ := cmd.Flags().GetDuration("timeout"); timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}

			cliClient := mustGetContextCLIClient(cmd)
			return cliClient.SyncDocuments(ctx, collectionID, docIDs)
		},
	}

	cmd.Flags().Duration("timeout", 0, "Timeout for sync operations")
	return cmd
}
