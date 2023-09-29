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
	"github.com/sourcenetwork/defradb/http"
)

func MakeDocumentKeysCommand() *cobra.Command {
	var collection string
	var cmd = &cobra.Command{
		Use:   "keys --collection <collection>",
		Short: "List all collection document keys.",
		Long: `List all collection document keys.
		
Example:
  defradb client document keys --collection User keys
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetStoreContext(cmd)

			col, err := store.GetCollectionByName(cmd.Context(), collection)
			if err != nil {
				return err
			}
			if tx, ok := cmd.Context().Value(txContextKey).(datastore.Txn); ok {
				col = col.WithTxn(tx)
			}
			docCh, err := col.GetAllDocKeys(cmd.Context())
			if err != nil {
				return err
			}
			for docKey := range docCh {
				results := &http.DocKeyResult{
					Key: docKey.Key.String(),
				}
				if docKey.Err != nil {
					results.Error = docKey.Err.Error()
				}
				writeJSON(cmd, results) //nolint:errcheck
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&collection, "collection", "c", "", "Collection name")
	return cmd
}
