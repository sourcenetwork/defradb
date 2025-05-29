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
	"github.com/sourcenetwork/immutable"
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
)

func MakeCollectionDescribeCommand() *cobra.Command {
	var name string
	var collectionID string
	var versionID string
	var getInactive bool
	var cmd = &cobra.Command{
		Use:   "describe",
		Short: "View collection version.",
		Long: `Introspect collection types.

Example: view all collections
  defradb client collection describe
		
Example: view collection by name
  defradb client collection describe --name User
		
Example: view collection by collection id
  defradb client collection describe --collection-id bae123
		
Example: view collection by version id. This will also return inactive collections
  defradb client collection describe --version-id bae123
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := mustGetContextClient(cmd)

			options := client.CollectionFetchOptions{}
			if versionID != "" {
				options.VersionID = immutable.Some(versionID)
			}
			if collectionID != "" {
				options.CollectionID = immutable.Some(collectionID)
			}
			if name != "" {
				options.Name = immutable.Some(name)
			}
			if getInactive {
				options.IncludeInactive = immutable.Some(getInactive)
			}

			cols, err := c.GetCollections(
				cmd.Context(),
				options,
			)
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
	cmd.Flags().StringVar(&name, "name", "", "Collection name")
	cmd.Flags().StringVar(&collectionID, "collection-id", "", "Collection P2P identifier")
	cmd.Flags().StringVar(&versionID, "version-id", "", "Collection version ID")
	cmd.Flags().BoolVar(&getInactive, "get-inactive", false, "Get inactive collections as well as active")
	return cmd
}
