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
	"strings"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeIndexAddCommand(ctx context.Context) *cobra.Command {
	var collectionArg string
	var nameArg string
	var fieldsArg []string
	var uniqueArg bool
	var cmd = &cobra.Command{
		Use:   "add -c --collection <collection> --fields <fields[:ASC|:DESC]> [-n --name <name>] [--unique]",
		Short: "Adds a secondary index on a collection's field(s)",
		Long: `Adds a secondary index on a collection's field(s).

The --name flag is optional. If not provided, a name will be generated automatically.
The --unique flag is optional. If provided, the index will be unique.
If no order is specified for the field, the default value will be "ASC"`,
		ValidArgs: []string{"collection", "fields", "name"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			var fields []client.IndexedFieldDescription

			for _, field := range fieldsArg {
				// For each field, parse it into a field name and ascension order, separated by a colon
				// If there is no colon, assume the ascension order is ASC by default
				const asc = "ASC"
				const desc = "DESC"
				parts := strings.Split(field, ":")
				fieldName := parts[0]
				order := asc
				if len(parts) == 2 {
					order = strings.ToUpper(parts[1])
					if order != asc && order != desc {
						return NewErrInvalidAscensionOrder(field)
					}
				} else if len(parts) > 2 {
					return NewErrInvalidIndexFieldDescription(field)
				}
				fields = append(fields, client.IndexedFieldDescription{
					Name:       fieldName,
					Descending: order == desc,
				})
			}

			desc := client.AddIndexRequest{
				Name:   nameArg,
				Fields: fields,
				Unique: uniqueArg,
			}
			colOpt := options.WithIdentity(options.GetCollectionByName(), identity.FromContext(cmd.Context()))
			col, err := cliClient.GetCollectionByName(cmd.Context(), collectionArg, colOpt)
			if err != nil {
				return err
			}

			indOpt := options.WithIdentity(options.AddCollectionIndex(), identity.FromContext(cmd.Context()))
			descWithID, err := col.AddIndex(cmd.Context(), desc, indOpt)
			if err != nil {
				return err
			}
			return writeJSON(cmd, descWithID)
		},
	}

	EmbedCLIExample(ctx, cmd, "add an index for 'Users' collection on 'name' field",
		`defradb client index add --collection Users --fields name`)

	EmbedCLIExample(ctx, cmd, "add a named index for 'Users' collection on 'name' field",
		`defradb client index add --collection Users --fields name --name UsersByName`)

	EmbedCLIExample(ctx, cmd, "add a unique index for 'Users' collection on 'name' and 'age'",
		`defradb client index add --collection Users --fields name:ASC,age:DESC --unique`)

	cmd.Flags().StringVarP(&collectionArg, "collection", "c", "", "Collection name")
	cmd.Flags().StringVarP(&nameArg, "name", "n", "", "Index name")
	cmd.Flags().StringSliceVar(&fieldsArg, "fields", []string{}, "Fields to index")
	cmd.Flags().BoolVarP(&uniqueArg, "unique", "u", false, "Make the index unique")

	return cmd
}
