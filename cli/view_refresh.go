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
	var collectionID string
	var versionID string
	var getInactive bool
	var cmd = &cobra.Command{
		Use:   "refresh",
		Short: "Refresh views.",
		Long: `Refresh views, executing the underlying query and LensVm transforms and
persisting the results.

View is refreshed as the current user, meaning the cached items will reflect that user's
permissions. Subsequent query requests to the view, regardless of user, will receive
items from that cache.

Example: refresh all views
  defradb client view refresh

Example: refresh views by name
  defradb client view refresh --name UserView

Example: refresh views by schema root id
  defradb client view refresh --schema bae123

Example: refresh views by version id. This will also return inactive views
  defradb client view refresh --version-id bae123
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

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

			return cliClient.RefreshViews(
				cmd.Context(),
				options,
			)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "View name")
	cmd.Flags().StringVar(&collectionID, "collection-id", "", "View collection ID")
	cmd.Flags().StringVar(&versionID, "version-id", "", "View version ID")
	cmd.Flags().BoolVar(&getInactive, "get-inactive", false, "Get inactive views as well as active")
	return cmd
}
