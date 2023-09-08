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
	"github.com/sourcenetwork/defradb/datastore"
)

func MakeDocumentGetCommand() *cobra.Command {
	var showDeleted bool
	var collection string
	var cmd = &cobra.Command{
		Use:   "get --collection <collection> <docKey> [--show-deleted]",
		Short: "View detailed document info.",
		Long: `View detailed document info.

Example:
  defradb client document get --collection User bae123
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := cmd.Context().Value(storeContextKey).(client.Store)

			col, err := store.GetCollectionByName(cmd.Context(), collection)
			if err != nil {
				return err
			}
			if tx, ok := cmd.Context().Value(txContextKey).(datastore.Txn); ok {
				col = col.WithTxn(tx)
			}
			docKey, err := client.NewDocKeyFromString(args[0])
			if err != nil {
				return err
			}
			doc, err := col.Get(cmd.Context(), docKey, showDeleted)
			if err != nil {
				return err
			}
			docMap, err := doc.ToMap()
			if err != nil {
				return err
			}
			return writeJSON(cmd, docMap)
		},
	}
	cmd.Flags().BoolVar(&showDeleted, "show-deleted", false, "Show deleted documents")
	cmd.Flags().StringVarP(&collection, "collection", "c", "", "Collection name")
	return cmd
}
