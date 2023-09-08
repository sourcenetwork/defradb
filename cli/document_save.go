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
	"github.com/sourcenetwork/defradb/errors"
)

func MakeDocumentSaveCommand() *cobra.Command {
	var collection string
	var key string
	var cmd = &cobra.Command{
		Use:   "save --collection <collection> --key <docKey> <document>",
		Short: "Create or update a document.",
		Long: `Create or update a document.
		
Example:
  defradb client document save --collection User --key bae-123 '{ "name": "Bob" }'
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

			docKey, err := client.NewDocKeyFromString(key)
			if err != nil {
				return err
			}
			doc, err := col.Get(cmd.Context(), docKey, true)
			if err == nil {
				err = doc.SetWithJSON([]byte(args[0]))
			} else if errors.Is(err, client.ErrDocumentNotFound) {
				doc, err = client.NewDocFromJSON([]byte(args[0]))
			}
			if err != nil {
				return err
			}
			return col.Save(cmd.Context(), doc)
		},
	}
	cmd.Flags().StringVarP(&collection, "collection", "c", "", "Collection name")
	cmd.Flags().StringVar(&key, "key", "", "Document key")
	return cmd
}
