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
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeIndexDeleteCommand(ctx context.Context) *cobra.Command {
	var collectionArg string
	var nameArg string
	var cmd = &cobra.Command{
		Use:       "delete -c --collection <collection> -n --name <name>",
		Short:     "Delete a collection's secondary index",
		Long:      `Delete a collection's secondary index.`,
		ValidArgs: []string{"collection", "name"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if nameArg == "" {
				return client.ErrIndexNameRequired
			}

			cliClient := mustGetContextCLIClient(cmd)

			colOpt := options.WithIdentity(options.GetCollectionByName(), identity.FromContext(cmd.Context()))
			col, err := cliClient.GetCollectionByName(cmd.Context(), collectionArg, colOpt)
			if err != nil {
				return err
			}
			opt := options.WithIdentity(options.CollectionDeleteIndex(), identity.FromContext(cmd.Context()))
			return col.DeleteIndex(cmd.Context(), nameArg, opt)
		},
	}

	EmbedCLIExample(ctx, cmd, "delete the index 'UsersByName' for 'Users' collection",
		`defradb client index delete --collection Users --name UsersByName`)

	cmd.Flags().StringVarP(&collectionArg, "collection", "c", "", "Collection name")
	cmd.Flags().StringVarP(&nameArg, "name", "n", "", "Index name")

	return cmd
}
