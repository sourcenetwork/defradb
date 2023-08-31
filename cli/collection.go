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

func MakeCollectionCommand() *cobra.Command {
	var name string
	var schemaID string
	var versionID string
	var cmd = &cobra.Command{
		Use:   "collection",
		Short: "View detailed collection info.",
		Long:  `View detailed collection info.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := cmd.Context().Value(storeContextKey).(client.Store)

			switch {
			case name != "":
				col, err := store.GetCollectionByName(cmd.Context(), name)
				if err != nil {
					return err
				}
				return writeJSON(cmd, col.Description())
			case schemaID != "":
				col, err := store.GetCollectionBySchemaID(cmd.Context(), schemaID)
				if err != nil {
					return err
				}
				return writeJSON(cmd, col.Description())
			case versionID != "":
				col, err := store.GetCollectionByVersionID(cmd.Context(), versionID)
				if err != nil {
					return err
				}
				return writeJSON(cmd, col.Description())
			default:
				cols, err := store.GetAllCollections(cmd.Context())
				if err != nil {
					return err
				}
				colDesc := make([]client.CollectionDescription, len(cols))
				for i, col := range cols {
					colDesc[i] = col.Description()
				}
				return writeJSON(cmd, colDesc)
			}
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Get collection by name")
	cmd.Flags().StringVar(&schemaID, "schema", "", "Get collection by schema ID")
	cmd.Flags().StringVar(&versionID, "version", "", "Get collection by version ID")
	return cmd
}
