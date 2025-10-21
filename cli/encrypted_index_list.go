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
)

func MakeEncryptedIndexListCommand() *cobra.Command {
	var collectionArg string
	var cmd = &cobra.Command{
		Use:   "list [-c --collection <collection>]",
		Short: "Lists the encrypted indexes in the database or for a specific collection",
		Long: `Shows the list encrypted indexes in the database or for a specific collection.
		
If the --collection flag is provided, only the encrypted indexes for that collection will be shown.
Otherwise, all encrypted indexes in the database will be shown.

Example: show all encrypted indexes for 'Users' collection:
  defradb client encrypted-index list --collection Users`,
		ValidArgs: []string{"collection"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			switch {
			case collectionArg != "":
				col, err := cliClient.GetCollectionByName(cmd.Context(), collectionArg)
				if err != nil {
					return err
				}
				indexes, err := col.ListEncryptedIndexes(cmd.Context())
				if err != nil {
					return err
				}
				return writeJSON(cmd, indexes)
			default:
				indexes, err := cliClient.ListAllEncryptedIndexes(cmd.Context())
				if err != nil {
					return err
				}
				return writeJSON(cmd, indexes)
			}
		},
	}
	cmd.Flags().StringVarP(&collectionArg, "collection", "c", "", "Collection name")

	return cmd
}
