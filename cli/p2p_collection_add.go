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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
)

func MakeP2PCollectionAddCommand(cfg *config.Config, db client.DB) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add [collectionID]",
		Short: "Add P2P collections",
		Long: `Add P2P collections to the synchronized pubsub topics.
The collections are synchronized between nodes of a pubsub network.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return errors.New("must specify collectionID")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return db.AddP2PCollection(cmd.Context(), args[0])
		},
	}
	return cmd
}
