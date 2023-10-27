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

func MakeSchemaDescribeCommand() *cobra.Command {
	var name string
	var root string
	var versionID string

	var cmd = &cobra.Command{
		Use:   "describe",
		Short: "View schema descriptions.",
		Long: `Introspect schema types.

Example: view all schemas
  defradb client schema describe
		
Example: view schemas by name
  defradb client schema describe --name User
		
Example: view schemas by root
  defradb client schema describe --root bae123
		
Example: view a single schema by version id
  defradb client schema describe --version bae123
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetStoreContext(cmd)

			var schemas []client.SchemaDescription
			switch {
			case versionID != "":
				schema, err := store.GetSchemaByVersionID(cmd.Context(), versionID)
				if err != nil {
					return err
				}
				return writeJSON(cmd, schema)

			case root != "":
				s, err := store.GetSchemasByRoot(cmd.Context(), root)
				if err != nil {
					return err
				}
				schemas = s

			case name != "":
				s, err := store.GetSchemasByName(cmd.Context(), name)
				if err != nil {
					return err
				}
				schemas = s

			default:
				s, err := store.GetAllSchemas(cmd.Context())
				if err != nil {
					return err
				}
				schemas = s
			}

			return writeJSON(cmd, schemas)
		},
	}
	cmd.PersistentFlags().StringVar(&name, "name", "", "Schema name")
	cmd.PersistentFlags().StringVar(&root, "root", "", "Schema root")
	cmd.PersistentFlags().StringVar(&versionID, "version", "", "Schema Version ID")
	return cmd
}
