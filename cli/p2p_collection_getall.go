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
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
)

func MakeP2PCollectionGetallCommand(cfg *config.Config, db client.DB) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "getall",
		Short: "Get all P2P collections",
		Long: `Get all P2P collections in the pubsub topics.
This is the list of collections of the node that are synchronized on the pubsub network.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.NoArgs(cmd, args); err != nil {
				return errors.New("must specify no argument")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cols, err := db.GetAllP2PCollections(cmd.Context())
			if err != nil {
				return err
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(cols)
		},
	}
	return cmd
}
