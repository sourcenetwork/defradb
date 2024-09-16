// Copyright 2024 Democratized Data Foundation
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

func MakeViewRefreshCommand() *cobra.Command {
	var name string
	var schemaRoot string
	var versionID string
	var getInactive bool
	var cmd = &cobra.Command{
		Use:   "refresh",
		Short: "Refresh views.",
		Long: `Refresh views, executing the underlying query and LensVm transforms and
persisting the results.

View is refreshed as the current user, meaning results returned for all subsequent query requests
to the view will recieve items accessible to the user refreshing the view's permissions.

Example: refresh all views
  defradb client view refresh

Example: refresh views by name
  defradb client view refresh --name UserView

Example: refresh views by schema root id
  defradb client view refresh --schema bae123

Example: refresh views by version id. This will also return inactive views
  defradb client view refresh --version bae123
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetContextStore(cmd)

			options := client.CollectionFetchOptions{}
			if versionID != "" {
				options.SchemaVersionID = immutable.Some(versionID)
			}
			if schemaRoot != "" {
				options.SchemaRoot = immutable.Some(schemaRoot)
			}
			if name != "" {
				options.Name = immutable.Some(name)
			}
			if getInactive {
				options.IncludeInactive = immutable.Some(getInactive)
			}

			return store.RefreshViews(
				cmd.Context(),
				options,
			)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "View name")
	cmd.Flags().StringVar(&schemaRoot, "schema", "", "View schema Root")
	cmd.Flags().StringVar(&versionID, "version", "", "View version ID")
	cmd.Flags().BoolVar(&getInactive, "get-inactive", false, "Get inactive views as well as active")
	return cmd
}
