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

func MakeP2PDocumentRemoveCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "remove [docIDs]",
		Short: "Remove P2P documents",
		Long: `Remove P2P documents from the followed pubsub topics.
The removed documents will no longer be synchronized between nodes.

Example: remove single document
  defradb client p2p document remove bae123

Example: remove multiple documents
  defradb client p2p document remove bae123,bae456
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

			return cliClient.RemoveP2PDocuments(cmd.Context(), collectionIDs...)
		},
	}
	return cmd
}
