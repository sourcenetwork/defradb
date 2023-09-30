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
)

func MakeCollectionDeleteCommand() *cobra.Command {
	var keys []string
	var filter string
	var cmd = &cobra.Command{
		Use:   "delete [--filter <filter> --key <key>]",
		Short: "Delete documents by key or filter.",
		Long: `Delete documents by key or filter and lists the number of documents deleted.
		
Example: delete by key(s)
  defradb client collection delete --name User --key bae-123,bae-456

Example: delete by filter
  defradb client collection delete --name User --filter '{ "_gte": { "points": 100 } }'
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			col, ok := cmd.Context().Value(colContextKey).(client.Collection)
			if !ok {
				return cmd.Usage()
			}

			switch {
			case len(keys) == 1:
				docKey, err := client.NewDocKeyFromString(keys[0])
				if err != nil {
					return err
				}
				res, err := col.DeleteWithKey(cmd.Context(), docKey)
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
				res, err := col.DeleteWithKeys(cmd.Context(), docKeys)
				if err != nil {
					return err
				}
				return writeJSON(cmd, res)
			case filter != "":
				res, err := col.DeleteWithFilter(cmd.Context(), filter)
				if err != nil {
					return err
				}
				return writeJSON(cmd, res)
			default:
				return fmt.Errorf("document key or filter must be defined")
			}
		},
	}
	cmd.Flags().StringSliceVar(&keys, "key", nil, "Document key")
	cmd.Flags().StringVar(&filter, "filter", "", "Document filter")
	return cmd
}
