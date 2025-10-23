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
	"context"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

func MakeViewRefreshCommand(ctx context.Context) *cobra.Command {
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
items from that cache.`,
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

	EmbedCLIExample(ctx, cmd, "refresh all views",
		`defradb client view refresh`)

	EmbedCLIExample(ctx, cmd, "refresh views by name",
		`defradb client view refresh --name UserView`)

	EmbedCLIExample(ctx, cmd, "refresh views by collection id",
		`defradb client view refresh --collection-id bae123`)

	EmbedCLIExample(ctx, cmd, "refresh views by version id",
		`defradb client view refresh --version-id bae123`)

	cmd.Flags().StringVar(&name, "name", "", "View name")
	cmd.Flags().StringVar(&collectionID, "collection-id", "", "View collection ID")
	cmd.Flags().StringVar(&versionID, "version-id", "", "View version ID")
	cmd.Flags().BoolVar(&getInactive, "get-inactive", false, "Get inactive views as well as active")
	return cmd
}
