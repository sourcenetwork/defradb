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
	"strings"

	"github.com/spf13/cobra"
)

func MakeP2PCollectionRemoveCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "remove [collectionIDs]",
		Short: "Remove P2P collections",
		Long: `Remove P2P collections from the followed pubsub topics.
The removed collections will no longer be synchronized between nodes.

Example: remove single collection
  defradb client p2p collection remove bae123

Example: remove multiple collections
  defradb client p2p collection remove bae123,bae456
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := mustGetContextClient(cmd)

			var collectionIDs []string
			for _, id := range strings.Split(args[0], ",") {
				id = strings.TrimSpace(id)
				if id == "" {
					continue
				}
				collectionIDs = append(collectionIDs, id)
			}

			return client.RemoveP2PCollections(cmd.Context(), collectionIDs...)
		},
	}
	return cmd
}
