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

func MakeP2PCollectionAddCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add [collectionIDs]",
		Short: "Add P2P collections",
		Long: `Add P2P collections to the synchronized pubsub topics.
The collections are synchronized between nodes of a pubsub network.

Example: add single collection
  defradb client p2p collection add bae123

Example: add multiple collections
  defradb client p2p collection add bae123,bae456
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetStoreContext(cmd)
			collectionIDs := strings.Split(args[0], ",")
			return store.AddP2PCollections(cmd.Context(), collectionIDs)
		},
	}
	return cmd
}
