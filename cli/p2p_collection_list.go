// Copyright 2022 Democratized Data Foundation
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

func MakeP2PCollectionListCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List P2P collections",
		Long: `List P2P collections in the pubsub topics.
This is the list of collections of the node that are synchronized on the pubsub network.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			opt := options.WithIdentity(options.ListP2PCollections(), identity.FromContext(cmd.Context()))
			cols, err := cliClient.ListP2PCollections(cmd.Context(), opt)
			if err != nil {
				return err
			}
			return writeJSON(cmd, cols)
		},
	}
	return cmd
}
