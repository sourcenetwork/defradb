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
	"strings"

	"github.com/spf13/cobra"
)

func MakeP2PCollectionDeleteCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "delete [collectionNames]",
		Short: "Delete P2P collections",
		Long: `Delete P2P collections from the followed pubsub topics.
The removed collections will no longer be synchronized between nodes.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			var collectionNames []string
			for _, id := range strings.Split(args[0], ",") {
				id = strings.TrimSpace(id)
				if id == "" {
					continue
				}
				collectionNames = append(collectionNames, id)
			}

			return cliClient.DeleteP2PCollections(cmd.Context(), collectionNames...)
		},
	}

	EmbedCLIExample(ctx, cmd, "delete single collection",
		`defradb client p2p collection delete User`)

	EmbedCLIExample(ctx, cmd, "delete multiple collections",
		`defradb client p2p collection delete User,Address`)

	return cmd
}
