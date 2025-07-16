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
	"strings"

	"github.com/spf13/cobra"
)

func MakeP2PDocumentAddCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add [docIDs]",
		Short: "Add P2P documents",
		Long: `Add P2P documents to the synchronized pubsub topics.
The documents are synchronized between nodes of a pubsub network.

Example: add single document
  defradb client p2p document add bae123

Example: add multiple documents
  defradb client p2p document add bae123,bae456
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			var collectionIDs []string
			for _, id := range strings.Split(args[0], ",") {
				id = strings.TrimSpace(id)
				if id == "" {
					continue
				}
				collectionIDs = append(collectionIDs, id)
			}

			return cliClient.AddP2PDocuments(cmd.Context(), collectionIDs...)
		},
	}
	return cmd
}
