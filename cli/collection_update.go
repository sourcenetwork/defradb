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
)

func MakeCollectionUpdateCommand() *cobra.Command {
	var argDocID string
	var filter string
	var updater string
	var cmd = &cobra.Command{
		Use:   "update [-i --identity] [--filter <filter> --docID <docID> --updater <updater>] <document>",
		Short: "Update documents by docID or filter.",
		Long: `Update documents by docID or filter.
		
Example: update from string:
  defradb client collection update --name User --docID bae-123 '{ "name": "Bob" }'

Example: update by filter:
  defradb client collection update --name User \
  --filter '{ "_gte": { "points": 100 } }' --updater '{ "verified": true }'

Example: update by docID:
  defradb client collection update --name User \
  --docID bae-123 --updater '{ "verified": true }'

Example: update private docID, with identity:
  defradb client collection update -i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f --name User \
  --docID bae-123 --updater '{ "verified": true }'
		`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			col, ok := tryGetContextCollection(cmd)
			if !ok {
				return cmd.Usage()
			}

			switch {
			case filter != "" || updater != "":
				var filterValue any
				if err := json.Unmarshal([]byte(filter), &filterValue); err != nil {
					return err
				}
				res, err := col.UpdateWithFilter(cmd.Context(), filterValue, updater)
				if err != nil {
					return err
				}
				return writeJSON(cmd, res)
			case argDocID != "" && len(args) == 1:
				docID, err := client.NewDocIDFromString(argDocID)
				if err != nil {
					return err
				}
				doc, err := col.Get(cmd.Context(), docID, true)
				if err != nil {
					return err
				}
				if err := doc.SetWithJSON([]byte(args[0])); err != nil {
					return err
				}
				return col.Update(cmd.Context(), doc)
			default:
				return ErrNoDocIDOrFilter
			}
		},
	}
	cmd.Flags().StringVar(&argDocID, "docID", "", "Document ID")
	cmd.Flags().StringVar(&filter, "filter", "", "Document filter")
	cmd.Flags().StringVar(&updater, "updater", "", "Document updater")
	return cmd
}
