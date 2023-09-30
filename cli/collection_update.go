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

func MakeCollectionUpdateCommand() *cobra.Command {
	var keys []string
	var filter string
	var updater string
	var cmd = &cobra.Command{
		Use:   "update [--filter <filter> --key <key> --updater <updater>] <document>",
		Short: "Update documents by key or filter.",
		Long: `Update documents by key or filter.
		
Example:
  defradb client collection update --name User --key bae-123 '{ "name": "Bob" }'

Example: update by filter
  defradb client collection update --name User \
  --filter '{ "_gte": { "points": 100 } }' --updater '{ "verified": true }'

Example: update by keys
  defradb client collection update --name User \
  --key bae-123,bae-456 --updater '{ "verified": true }'
		`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			col, ok := cmd.Context().Value(colContextKey).(client.Collection)
			if !ok {
				return cmd.Usage()
			}

			switch {
			case len(keys) == 1 && updater != "":
				docKey, err := client.NewDocKeyFromString(keys[0])
				if err != nil {
					return err
				}
				res, err := col.UpdateWithKey(cmd.Context(), docKey, updater)
				if err != nil {
					return err
				}
				return writeJSON(cmd, res)
			case len(keys) > 1 && updater != "":
				docKeys := make([]client.DocKey, len(keys))
				for i, v := range keys {
					docKey, err := client.NewDocKeyFromString(v)
					if err != nil {
						return err
					}
					docKeys[i] = docKey
				}
				res, err := col.UpdateWithKeys(cmd.Context(), docKeys, updater)
				if err != nil {
					return err
				}
				return writeJSON(cmd, res)
			case filter != "" && updater != "":
				res, err := col.UpdateWithFilter(cmd.Context(), filter, updater)
				if err != nil {
					return err
				}
				return writeJSON(cmd, res)
			case len(keys) == 1 && len(args) == 1:
				docKey, err := client.NewDocKeyFromString(keys[0])
				if err != nil {
					return err
				}
				doc, err := col.Get(cmd.Context(), docKey, true)
				if err != nil {
					return err
				}
				if err := doc.SetWithJSON([]byte(args[0])); err != nil {
					return err
				}
				return col.Update(cmd.Context(), doc)
			default:
				return fmt.Errorf("document key or filter must be defined")
			}
		},
	}
	cmd.Flags().StringSliceVar(&keys, "key", nil, "Document key")
	cmd.Flags().StringVar(&filter, "filter", "", "Document filter")
	cmd.Flags().StringVar(&updater, "updater", "", "Document updater")
	return cmd
}
