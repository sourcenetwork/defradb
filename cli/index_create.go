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
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
)

func MakeIndexCreateCommand(cfg *config.Config, db client.DB) *cobra.Command {
	var collectionArg string
	var nameArg string
	var fieldsArg []string
	var cmd = &cobra.Command{
		Use:   "create -c --collection <collection> --fields <fields> [-n --name <name>]",
		Short: "Creates a secondary index on a collection's field(s)",
		Long: `Creates a secondary index on a collection's field(s).
		
The --name flag is optional. If not provided, a name will be generated automatically.

Example: create an index for 'Users' collection on 'name' field:
  defradb client index create --collection Users --fields name

Example: create a named index for 'Users' collection on 'name' field:
  defradb client index create --collection Users --fields name --name UsersByName`,
		ValidArgs: []string{"collection", "fields", "name"},
		RunE: func(cmd *cobra.Command, args []string) error {
			var fields []client.IndexedFieldDescription
			for _, name := range fieldsArg {
				fields = append(fields, client.IndexedFieldDescription{Name: name})
			}
			desc := client.IndexDescription{
				Name:   nameArg,
				Fields: fields,
			}
			col, err := db.GetCollectionByName(cmd.Context(), collectionArg)
			if err != nil {
				return err
			}
			desc, err = col.CreateIndex(cmd.Context(), desc)
			if err != nil {
				return err
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(desc)
		},
	}
	cmd.Flags().StringVarP(&collectionArg, "collection", "c", "", "Collection name")
	cmd.Flags().StringVarP(&nameArg, "name", "n", "", "Index name")
	cmd.Flags().StringSliceVar(&fieldsArg, "fields", []string{}, "Fields to index")

	return cmd
}
