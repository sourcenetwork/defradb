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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

func MakeEncryptedIndexAddCommand(ctx context.Context) *cobra.Command {
	var collectionArg string
	var fieldArg string
	var typeArg string
	var cmd = &cobra.Command{
		Use:   "add -c --collection <collection> --field <field> [--type <type>]",
		Short: "Adds an encrypted index on a collection's field",
		Long: `Adds an encrypted index on a collection's field.

The --type flag is optional. If not provided, the default value will be "equality".

Currently only "equality" type is supported.`,
		ValidArgs: []string{"collection", "field", "type"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			addReq := client.EncryptedIndexDescription{
				FieldName: fieldArg,
				Type:      client.EncryptedIndexType(typeArg),
			}
			opt := options.WithIdentity(options.GetCollectionByName(), iIdentity.FromContext(cmd.Context()))
			col, err := cliClient.GetCollectionByName(cmd.Context(), collectionArg, opt)
			if err != nil {
				return err
			}

			addOpt := options.WithIdentity(options.AddEncryptedIndex(), iIdentity.FromContext(cmd.Context()))
			descWithID, err := col.AddEncryptedIndex(cmd.Context(), addReq, addOpt)
			if err != nil {
				return err
			}
			return writeJSON(cmd, descWithID)
		},
	}

	EmbedCLIExample(ctx, cmd, "add an index for 'Users' collection on 'name' field",
		`defradb client encrypted-index add --collection Users --field name`)

	cmd.Flags().StringVarP(&collectionArg, "collection", "c", "", "Collection name")
	cmd.Flags().StringVar(&fieldArg, "field", "", "Field to index")
	cmd.Flags().StringVar(&typeArg, "type", "", "Type of index to add")

	return cmd
}
