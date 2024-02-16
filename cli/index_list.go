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

	"github.com/sourcenetwork/defradb/datastore"
)

func MakeIndexListCommand() *cobra.Command {
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
			store := mustGetContextStore(cmd)

			switch {
			case collectionArg != "":
				col, err := store.GetCollectionByName(cmd.Context(), collectionArg)
				if err != nil {
					return err
				}
				if tx, ok := cmd.Context().Value(txContextKey).(datastore.Txn); ok {
					col = col.WithTxn(tx)
				}
				indexes, err := col.GetIndexes(cmd.Context())
				if err != nil {
					return err
				}
				return writeJSON(cmd, indexes)
			default:
				indexes, err := store.GetAllIndexes(cmd.Context())
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
