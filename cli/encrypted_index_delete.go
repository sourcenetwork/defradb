// Copyright 2025 Democratized Data Foundation
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

func MakeEncryptedIndexDeleteCommand(ctx context.Context) *cobra.Command {
	var collectionArg string
	var fieldArg string
	var cmd = &cobra.Command{
		Use:       "delete -c --collection <collection> --field <field>",
		Short:     "Delete an encrypted index from a collection's field",
		Long:      `Delete an encrypted index from a collection's field.`,
		ValidArgs: []string{"collection", "field"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			opt := options.WithIdentity(options.GetCollectionByName(), iIdentity.FromContext(cmd.Context()))
			col, err := cliClient.GetCollectionByName(cmd.Context(), collectionArg, opt)
			if err != nil {
				return err
			}

			deleteOpt := options.WithIdentity(options.DeleteEncryptedIndex(), iIdentity.FromContext(cmd.Context()))
			return col.DeleteEncryptedIndex(cmd.Context(), fieldArg, deleteOpt)
		},
	}

	EmbedCLIExample(ctx, cmd, "delete an encrypted index for 'Users' collection on 'name' field",
		`defradb client encrypted-index delete --collection Users --field name`)

	cmd.Flags().StringVarP(&collectionArg, "collection", "c", "", "Collection name")
	cmd.Flags().StringVar(&fieldArg, "field", "", "Field name to delete encrypted index from")

	return cmd
}
