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

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeViewRefreshCommand(ctx context.Context) *cobra.Command {
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

			name, _ := cmd.Flags().GetString("collection-name")
			collectionID, _ := cmd.Flags().GetString("collection-id")
			versionID, _ := cmd.Flags().GetString("version-id")
			getInactive, _ := cmd.Flags().GetBool("get-inactive")

			opt := options.WithIdentity(options.RefreshViews(), identity.FromContext(cmd.Context()))
			if versionID != "" {
				opt.SetVersionID(versionID)
			}
			if collectionID != "" {
				opt.SetCollectionID(collectionID)
			}
			if name != "" {
				opt.SetCollectionName(name)
			}
			if getInactive {
				opt.SetGetInactive(getInactive)
			}

			return cliClient.RefreshViews(
				cmd.Context(),
				opt,
			)
		},
	}

	EmbedCLIExample(ctx, cmd, "refresh all views",
		`defradb client view refresh`)

	EmbedCLIExample(ctx, cmd, "refresh views by name",
		`defradb client view refresh --collection-name UserView`)

	EmbedCLIExample(ctx, cmd, "refresh views by collection id",
		`defradb client view refresh --collection-id bae123`)

	EmbedCLIExample(ctx, cmd, "refresh views by version id",
		`defradb client view refresh --version-id bae123`)

	setCollectionSelectorFlags(cmd)
	return cmd
}
