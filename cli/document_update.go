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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

func MakeDocumentUpdateCommand() *cobra.Command {
	var collection string
	var keys []string
	var filter string
	var cmd = &cobra.Command{
		Use:   "update --collection <collection> [--filter <filter> --key <key>] <updater>",
		Short: "Update documents by key or filter.",
		Long:  `Update documents by key or filter`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := cmd.Context().Value(storeContextKey).(client.Store)

			col, err := store.GetCollectionByName(cmd.Context(), collection)
			if err != nil {
				return err
			}
			if tx, ok := cmd.Context().Value(txContextKey).(datastore.Txn); ok {
				col = col.WithTxn(tx)
			}

			switch {
			case len(keys) == 1:
				docKey, err := client.NewDocKeyFromString(keys[0])
				if err != nil {
					return err
				}
				res, err := col.UpdateWithKey(cmd.Context(), docKey, args[0])
				if err != nil {
					return err
				}
				return writeJSON(cmd, res)
			case len(keys) > 1:
				docKeys := make([]client.DocKey, len(keys))
				for i, v := range keys {
					docKey, err := client.NewDocKeyFromString(v)
					if err != nil {
						return err
					}
					docKeys[i] = docKey
				}
				res, err := col.UpdateWithKeys(cmd.Context(), docKeys, args[0])
				if err != nil {
					return err
				}
				return writeJSON(cmd, res)
			case filter != "":
				res, err := col.UpdateWithFilter(cmd.Context(), filter, args[0])
				if err != nil {
					return err
				}
				return writeJSON(cmd, res)
			default:
				return fmt.Errorf("document key or filter must be defined")
			}
		},
	}
	cmd.Flags().StringVarP(&collection, "collection", "c", "", "Collection name")
	cmd.Flags().StringSliceVar(&keys, "key", nil, "Document key")
	cmd.Flags().StringVar(&filter, "filter", "", "Document filter")
	return cmd
}
