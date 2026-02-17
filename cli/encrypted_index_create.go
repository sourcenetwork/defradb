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

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

func MakeEncryptedIndexCreateCommand(ctx context.Context) *cobra.Command {
	var collectionArg string
	var fieldArg string
	var typeArg string
	var cmd = &cobra.Command{
		Use:   "create -c --collection <collection> --field <field> [--type <type>]",
		Short: "Creates an encrypted index on a collection's field",
		Long: `Creates an encrypted index on a collection's field.

The --type flag is optional. If not provided, the default value will be "equality".

Currently only "equality" type is supported.`,
		ValidArgs: []string{"collection", "field", "type"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			createReq := client.EncryptedIndexDescription{
				FieldName: fieldArg,
				Type:      client.EncryptedIndexType(typeArg),
			}
			opt := options.WithIdentity(options.GetCollectionByName(), acpIdentity.FromContext(cmd.Context()))
			col, err := cliClient.GetCollectionByName(cmd.Context(), collectionArg, opt)
			if err != nil {
				return err
			}

			createOpt := options.WithIdentity(options.CreateEncryptedIndex(), acpIdentity.FromContext(cmd.Context()))
			descWithID, err := col.CreateEncryptedIndex(cmd.Context(), createReq, createOpt)
			if err != nil {
				return err
			}
			return writeJSON(cmd, descWithID)
		},
	}

	EmbedCLIExample(ctx, cmd, "create an index for 'Users' collection on 'name' field",
		`defradb client encrypted-index create --collection Users --field name`)

	cmd.Flags().StringVarP(&collectionArg, "collection", "c", "", "Collection name")
	cmd.Flags().StringVar(&fieldArg, "field", "", "Field to index")
	cmd.Flags().StringVar(&typeArg, "type", "", "Type of index to create")

	return cmd
}
