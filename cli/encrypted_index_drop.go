// Copyright 2025 Democratized Data Foundation
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
)

func MakeEncryptedIndexDropCommand() *cobra.Command {
	var collectionArg string
	var fieldArg string
	var cmd = &cobra.Command{
		Use:   "drop -c --collection <collection> --field <field>",
		Short: "Drop an encrypted index from a collection's field",
		Long: `Drop an encrypted index from a collection's field.

Example: drop an encrypted index for 'Users' collection on 'name' field:
  defradb client encrypted-index drop --collection Users --field name
`,
		ValidArgs: []string{"collection", "field"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			col, err := cliClient.GetCollectionByName(cmd.Context(), collectionArg)
			if err != nil {
				return err
			}
			
			return col.DropEncryptedIndex(cmd.Context(), fieldArg)
		},
	}
	cmd.Flags().StringVarP(&collectionArg, "collection", "c", "", "Collection name")
	cmd.Flags().StringVar(&fieldArg, "field", "", "Field name to drop encrypted index from")

	return cmd
}