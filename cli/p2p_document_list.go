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

func MakeP2PDocumentListCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List P2P documents",
		Long: `List all P2P documents in the pubsub topics.
This is the list of documents of the node that are synchronized on the pubsub network.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			opt := options.WithIdentity(options.ListP2PDocuments(), identity.FromContext(cmd.Context()))
			cols, err := cliClient.ListP2PDocuments(cmd.Context(), opt)
			if err != nil {
				return err
			}
			return writeJSON(cmd, cols)
		},
	}
	return cmd
}
