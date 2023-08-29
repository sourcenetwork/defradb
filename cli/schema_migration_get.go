// Copyright 2023 Democratized Data Foundation
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
)

func MakeSchemaMigrationGetCommand(cfg *config.Config, db client.DB) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get",
		Short: "Gets the schema migrations within DefraDB",
		Long: `Gets the schema migrations within the local DefraDB node.

Example:
  defradb client schema migration get'

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cfgs, err := db.LensRegistry().Config(cmd.Context())
			if err != nil {
				return err
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(cfgs)
		},
	}
	return cmd
}
