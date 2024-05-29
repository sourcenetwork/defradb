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
)

func MakeCollectionDeleteCommand() *cobra.Command {
	var argDocID string
	var filter string
	var cmd = &cobra.Command{
		Use:   "delete [-i --identity] [--filter <filter> --docID <docID>]",
		Short: "Delete documents by docID or filter.",
		Long: `Delete documents by docID or filter and lists the number of documents deleted.
		
Example: delete by docID:
  defradb client collection delete  --name User --docID bae-123

Example: delete by docID with identity:
  defradb client collection delete --name User --docID bae-123 \
  	-i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f

Example: delete by filter:
  defradb client collection delete --name User --filter '{ "_gte": { "points": 100 } }'
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			col, ok := tryGetContextCollection(cmd)
			if !ok {
				return cmd.Usage()
			}

			switch {
			case argDocID != "":
				docID, err := client.NewDocIDFromString(argDocID)
				if err != nil {
					return err
				}
				_, err = col.Delete(cmd.Context(), docID)
				return err
			case filter != "":
				res, err := col.DeleteWithFilter(cmd.Context(), filter)
				if err != nil {
					return err
				}
				return writeJSON(cmd, res)
			default:
				return ErrNoDocIDOrFilter
			}
		},
	}
	cmd.Flags().StringVar(&argDocID, "docID", "", "Document ID")
	cmd.Flags().StringVar(&filter, "filter", "", "Document filter")
	return cmd
}
