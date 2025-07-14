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
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
)

func MakeP2PSyncCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sync",
		Short: "P2P document synchronization commands",
		Long:  "P2P document synchronization commands",
	}

	return cmd
}

func MakeP2PSyncDocumentsCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "docs [collection-id] [doc-id...]",
		Short:   "Synchronize specific documents from the network",
		Aliases: []string{"documents"},
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			collectionID := args[0]
			docIDs := args[1:]

			var opts []client.DocSyncOption
			if timeout, _ := cmd.Flags().GetDuration("timeout"); timeout > 0 {
				opts = append(opts, client.DocSyncWithTimeout(timeout))
			}

			cliClient := mustGetContextCLIClient(cmd)
			return <-cliClient.SyncDocuments(cmd.Context(), collectionID, docIDs, opts...)
		},
	}

	cmd.Flags().Duration("timeout", 0, "Timeout for sync operations")
	return cmd
}
