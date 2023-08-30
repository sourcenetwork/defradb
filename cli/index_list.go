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
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
)

func MakeIndexListCommand(cfg *config.Config) *cobra.Command {
	var collectionArg string
	var cmd = &cobra.Command{
		Use:   "list [-c --collection <collection>]",
		Short: "Shows the list indexes in the database or for a specific collection",
		Long: `Shows the list indexes in the database or for a specific collection.
		
If the --collection flag is provided, only the indexes for that collection will be shown.
Otherwise, all indexes in the database will be shown.

Example: show all index for 'Users' collection:
  defradb client index list --collection Users`,
		ValidArgs: []string{"collection"},
		RunE: func(cmd *cobra.Command, args []string) error {
			store := cmd.Context().Value(storeContextKey).(client.Store)

			switch {
			case collectionArg != "":
				col, err := store.GetCollectionByName(cmd.Context(), collectionArg)
				if err != nil {
					return err
				}
				cols, err := col.GetIndexes(cmd.Context())
				if err != nil {
					return err
				}
				return writeJSON(cmd, cols)
			default:
				cols, err := store.GetAllIndexes(cmd.Context())
				if err != nil {
					return err
				}
				return writeJSON(cmd, cols)
			}
		},
	}
	cmd.Flags().StringVarP(&collectionArg, "collection", "c", "", "Collection name")

	return cmd
}
