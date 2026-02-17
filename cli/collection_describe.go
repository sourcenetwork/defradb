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
	"context"

	"github.com/spf13/cobra"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

func MakeCollectionDescribeCommand(ctx context.Context) *cobra.Command {
	var name string
	var collectionID string
	var versionID string
	var getInactive bool
	var cmd = &cobra.Command{
		Use:   "describe",
		Short: "View collection version.",
		Long:  `Introspect collection types.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			opt := options.WithIdentity(options.GetCollections(), acpIdentity.FromContext(cmd.Context()))
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

			cols, err := cliClient.GetCollections(
				cmd.Context(),
				opt,
			)
			if err != nil {
				return err
			}
			colDesc := make([]client.CollectionVersion, len(cols))
			for i, col := range cols {
				colDesc[i] = col.Version()
			}
			return writeJSON(cmd, colDesc)
		},
	}

	EmbedCLIExample(ctx, cmd, "view all collections",
		`defradb client collection describe`)

	EmbedCLIExample(ctx, cmd, "view collection by name",
		`defradb client collection describe --name User`)

	EmbedCLIExample(ctx, cmd, "view collection by collection id",
		`defradb client collection describe --collection-id bae123`)

	EmbedCLIExample(ctx, cmd, "view collection by version id",
		`defradb client collection describe --version-id bae123`)

	cmd.Flags().StringVar(&name, "name", "", "Collection name")
	cmd.Flags().StringVar(&collectionID, "collection-id", "", "Collection P2P identifier")
	cmd.Flags().StringVar(&versionID, "version-id", "", "Collection version ID")
	cmd.Flags().BoolVar(&getInactive, "get-inactive", false, "Get inactive collections as well as active")
	return cmd
}
