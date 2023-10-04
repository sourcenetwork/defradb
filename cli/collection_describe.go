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

type CollectionDefinition struct {
	Description client.CollectionDescription `json:"description"`
	Schema      client.SchemaDescription     `json:"schema"`
}

func MakeCollectionDescribeCommand() *cobra.Command {
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
			store := mustGetStoreContext(cmd)

			col, ok := tryGetCollectionContext(cmd)
			if ok {
				return writeJSON(cmd, CollectionDefinition{
					Description: col.Description(),
					Schema:      col.Schema(),
				})
			}
			// if no collection specified list all collections
			cols, err := store.GetAllCollections(cmd.Context())
			if err != nil {
				return err
			}
			colDesc := make([]CollectionDefinition, len(cols))
			for i, col := range cols {
				colDesc[i] = CollectionDefinition{
					Description: col.Description(),
					Schema:      col.Schema(),
				}
			}
			return writeJSON(cmd, colDesc)
		},
	}
	return cmd
}
