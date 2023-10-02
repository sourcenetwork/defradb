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
	"github.com/spf13/cobra"
)

func MakeP2PCollectionGetAllCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "getall",
		Short: "Get all P2P collections",
		Long: `Get all P2P collections in the pubsub topics.
This is the list of collections of the node that are synchronized on the pubsub network.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetStoreContext(cmd)

			cols, err := store.GetAllP2PCollections(cmd.Context())
			if err != nil {
				return err
			}
			return writeJSON(cmd, cols)
		},
	}
	return cmd
}
