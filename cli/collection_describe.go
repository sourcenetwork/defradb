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

	"github.com/sourcenetwork/defradb/client"
)

func MakeCollectionDescribeCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "describe",
		Short: "View collection version.",
		Long:  `Introspect collection types.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			opt := getCollectionSelectorOptions(cmd)
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
		`defradb client collection describe --collection-name User`)

	EmbedCLIExample(ctx, cmd, "view collection by collection id",
		`defradb client collection describe --collection-id bae123`)

	EmbedCLIExample(ctx, cmd, "view collection by version id",
		`defradb client collection describe --version-id bae123`)

	setCollectionSelectorFlags(cmd)
	return cmd
}
