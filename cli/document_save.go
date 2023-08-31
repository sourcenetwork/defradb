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
	"github.com/sourcenetwork/defradb/datastore"
)

func MakeDocumentSaveCommand() *cobra.Command {
	var collection string
	var cmd = &cobra.Command{
		Use:   "save --collection <collection> <document>",
		Short: "Create or update a docment.",
		Long:  `Create or update a docment.`,
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

			var docMap map[string]any
			if err := json.Unmarshal([]byte(args[0]), &docMap); err != nil {
				return err
			}
			doc, err := client.NewDocFromMap(docMap)
			if err != nil {
				return err
			}
			return col.Save(cmd.Context(), doc)
		},
	}
	cmd.Flags().StringVarP(&collection, "collection", "c", "", "Collection name")
	return cmd
}
