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

func MakeCollectionGetCommand() *cobra.Command {
	var showDeleted bool
	var cmd = &cobra.Command{
		Use:   "get [-i --identity] [--show-deleted] <docID> ",
		Short: "View document fields.",
		Long: `View document fields.

Example:
  defradb client collection get --name User bae-123

Example to get a private document we must use an identity:
  defradb client collection get -i cosmos1f2djr7dl9vhrk3twt3xwqp09nhtzec9mdkf70j --name User bae-123
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			col, ok := tryGetContextCollection(cmd)
			if !ok {
				return cmd.Usage()
			}

			docID, err := client.NewDocIDFromString(args[0])
			if err != nil {
				return err
			}
			doc, err := col.Get(cmd.Context(), docID, showDeleted)
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
	return cmd
}
