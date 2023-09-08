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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

func MakeDocumentCreateCommand() *cobra.Command {
	var collection string
	var cmd = &cobra.Command{
		Use:   "create --collection <collection> <document>",
		Short: "Create a new document.",
		Long: `Create a new document.

Example: create document
  defradb client collection create --collection User '{ "name": "Bob" }'

Example: create documents
  defradb client collection create --collection User '[{ "name": "Alice" }, { "name": "Bob" }]'
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

			var docMap any
			if err := json.Unmarshal([]byte(args[0]), &docMap); err != nil {
				return err
			}

			switch t := docMap.(type) {
			case map[string]any:
				doc, err := client.NewDocFromMap(t)
				if err != nil {
					return err
				}
				return col.Create(cmd.Context(), doc)
			case []map[string]any:
				docs := make([]*client.Document, len(t))
				for i, v := range t {
					doc, err := client.NewDocFromMap(v)
					if err != nil {
						return err
					}
					docs[i] = doc
				}
				return col.CreateMany(cmd.Context(), docs)
			default:
				return fmt.Errorf("invalid document")
			}
		},
	}
	cmd.Flags().StringVarP(&collection, "collection", "c", "", "Collection name")
	return cmd
}
