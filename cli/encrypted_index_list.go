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

	"github.com/sourcenetwork/defradb/client/options"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

func MakeEncryptedIndexListCommand(ctx context.Context) *cobra.Command {
	var collectionArg string
	var cmd = &cobra.Command{
		Use:   "list [-c --collection <collection>]",
		Short: "Lists the encrypted indexes in the database or for a specific collection",
		Long: `Shows the list encrypted indexes in the database or for a specific collection.

If the --collection flag is provided, only the encrypted indexes for that collection will be shown.
Otherwise, all encrypted indexes in the database will be shown.`,
		ValidArgs: []string{"collection"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			switch {
			case collectionArg != "":
				getColOpt := options.WithIdentity(options.GetCollectionByName(), iIdentity.FromContext(cmd.Context()))
				col, err := cliClient.GetCollectionByName(cmd.Context(), collectionArg, getColOpt)
				if err != nil {
					return err
				}
				listOpt := options.WithIdentity(options.ListCollectionEncryptedIndexes(), iIdentity.FromContext(cmd.Context()))
				indexes, err := col.ListEncryptedIndexes(cmd.Context(), listOpt)
				if err != nil {
					return err
				}
				return writeJSON(cmd, indexes)
			default:
				opt := options.WithIdentity(options.ListAllEncryptedIndexes(), iIdentity.FromContext(cmd.Context()))
				indexes, err := cliClient.ListAllEncryptedIndexes(cmd.Context(), opt)
				if err != nil {
					return err
				}
				return writeJSON(cmd, indexes)
			}
		},
	}

	EmbedCLIExample(ctx, cmd, "show all encrypted indexes for 'Users' collection",
		`defradb client encrypted-index list --collection Users`)

	cmd.Flags().StringVarP(&collectionArg, "collection", "c", "", "Collection name")

	return cmd
}
