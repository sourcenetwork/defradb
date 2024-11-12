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
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	
	"strings"
	
)

func MakeIndexCreateCommand() *cobra.Command {
	var collectionArg string
	var nameArg string
	var fieldsArg []string
	var uniqueArg bool
	var cmd = &cobra.Command{
		Use:   "create -c --collection <collection> --fields <fields> [-n --name <name>] [--unique]",
		Short: "Creates a secondary index on a collection's field(s)",
		Long: `Creates a secondary index on a collection's field(s).
		
The --name flag is optional. If not provided, a name will be generated automatically.
The --unique flag is optional. If provided, the index will be unique.

Example: create an index for 'Users' collection on 'name' field:
  defradb client index create --collection Users --fields name

Example: create a named index for 'Users' collection on 'name' field:
  defradb client index create --collection Users --fields name --name UsersByName`,
		ValidArgs: []string{"collection", "fields", "name"},
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetContextStore(cmd)

			var fields []client.IndexedFieldDescription
			
			for _, field := range fieldsArg {
				var fieldName string
				var order string

				// For each field, parse it into a field name and ascension order, separated by a colon
				// If there is no colon, assume the ascension order is ASC by default
				parts := strings.Split(field, ":")
				if len(parts) == 1 {
					fieldName = parts[0]
					order = "ASC"
				} else if len(parts) > 2 {
					return NewErrInvalidAscensionOrder(field)
				} else {
					fieldName = parts[0]
					order = strings.ToUpper(parts[1])
					if order != "ASC" && order != "DESC" {
						return NewErrInvalidAscensionOrder(field)
					}
				}
				fields = append(fields, client.IndexedFieldDescription{
					Name:       fieldName,
					Descending: order == "DESC",
				})
			}

			desc := client.IndexDescription{
				Name:   nameArg,
				Fields: fields,
				Unique: uniqueArg,
			}
			col, err := store.GetCollectionByName(cmd.Context(), collectionArg)
			if err != nil {
				return err
			}
			
			desc, err = col.CreateIndex(cmd.Context(), desc)
			if err != nil {
				return err
			}
			return writeJSON(cmd, desc)
		},
	}
	cmd.Flags().StringVarP(&collectionArg, "collection", "c", "", "Collection name")
	cmd.Flags().StringVarP(&nameArg, "name", "n", "", "Index name")
	cmd.Flags().StringSliceVar(&fieldsArg, "fields", []string{}, "Fields to index")
	cmd.Flags().BoolVarP(&uniqueArg, "unique", "u", false, "Make the index unique")

	return cmd
}
