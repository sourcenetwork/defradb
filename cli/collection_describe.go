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

func MakeCollectionDescribeCommand() *cobra.Command {
	var getInactive bool
	var cmd = &cobra.Command{
		Use:   "describe",
		Short: "View collection description.",
		Long: `Introspect collection types.

Example: view all collections
  defradb client collection describe
		
Example: view collection by name
  defradb client collection describe --name User
		
Example: view collection by schema id
  defradb client collection describe --schema bae123
		
Example: view collection by version id
  defradb client collection describe --version bae123
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetContextStore(cmd)

			col, ok := tryGetContextCollection(cmd)
			if ok {
				return writeJSON(cmd, col.Definition())
			}
			// if no collection specified list all collections
			cols, err := store.GetAllCollections(cmd.Context(), getInactive)
			if err != nil {
				return err
			}
			colDesc := make([]client.CollectionDefinition, len(cols))
			for i, col := range cols {
				colDesc[i] = col.Definition()
			}
			return writeJSON(cmd, colDesc)
		},
	}
	cmd.Flags().BoolVar(&getInactive, "get-inactive", false, "Get inactive collections as well as active")
	return cmd
}
