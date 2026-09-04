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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeIndexNewCommand(ctx context.Context) *cobra.Command {
	var collectionArg string
	var nameArg string
	var fieldsArg []string
	var uniqueArg bool
	var vectorArg string
	var cmd = &cobra.Command{
		Use:   "new",
		Short: "Make a new index on a collection's field(s)",
		Long: fmt.Sprintf(`Make a new index on a collection's field(s).

The --name flag is optional. If not provided, a name will be generated automatically.
The --unique flag is optional. If provided, the index will be unique.
If no order is specified for the field, the default value will be "ASC"

The --vector flag makes a vector index (on a single field, never unique). Its value is the index
config as JSON. Give the config for the algorithm you want under its own key; HNSW is the only one
today, e.g. '{"Metric":"COSINE","Dimensions":3,"HNSW":{}}'. Metric and Dimensions are the essentials
(Metric is the distance metric, one of COSINE, EUCLIDEAN or DOT; Dimensions is the vector length).
The metric cannot be changed later without dropping and recreating the index. The HNSW tuning
params are optional and default if omitted:
  M               links per node (higher = better recall, more memory and slower build); default %d
  EfConstruction  build-time search width (higher = better graph, slower build); default %d
  EfSearch        query-time search width (higher = better recall, slower queries); default %d

The index is built in the background. This command returns once the index is recorded, before
existing documents are indexed. The index starts "building" and becomes "ready" once complete, or
"failed" if it cannot be built. Use 'index list' to check its status.`,
			client.DefaultHNSWM, client.DefaultHNSWEfConstruction, client.DefaultHNSWEfSearch),
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

			desc := client.NewIndexRequest{
				Name:   nameArg,
				Fields: fields,
			}
			if vectorArg != "" {
				var vectorDesc client.VectorIndexDescription
				if err := json.Unmarshal([]byte(vectorArg), &vectorDesc); err != nil {
					return NewErrInvalidVectorIndexConfig(err)
				}
				desc.Vector = &vectorDesc
				// No ordered config to carry it, so the db rejects the combination.
				desc.Unique = uniqueArg
			} else {
				desc.Ordered = &client.OrderedIndexDescription{Unique: uniqueArg}
			}
			colOpt := options.WithIdentity(options.GetCollectionByName(), identity.FromContext(cmd.Context()))
			col, err := cliClient.GetCollectionByName(cmd.Context(), collectionArg, colOpt)
			if err != nil {
				return err
			}

			indOpt := options.WithIdentity(options.NewCollectionIndex(), identity.FromContext(cmd.Context()))
			descWithID, err := col.NewIndex(cmd.Context(), desc, indOpt)
			if err != nil {
				return err
			}
			return writeJSON(cmd, descWithID)
		},
	}

	EmbedCLIExample(ctx, cmd, "make a new index for 'Users' collection on 'name' field",
		`defradb client index new --collection Users --fields name`)

	EmbedCLIExample(ctx, cmd, "make a new named index for 'Users' collection on 'name' field",
		`defradb client index new --collection Users --fields name --name UsersByName`)

	EmbedCLIExample(ctx, cmd, "make a new unique index for 'Users' collection on 'name' and 'age'",
		`defradb client index new --collection Users --fields name:ASC,age:DESC --unique`)

	EmbedCLIExample(ctx, cmd, "make a new vector index for 'Users' collection on 'vec' field (HNSW defaults)",
		`defradb client index new --collection Users --fields vec `+
			`--vector '{"Metric":"COSINE","Dimensions":3,"HNSW":{}}'`)

	EmbedCLIExample(ctx, cmd, "make a new vector index tuning the HNSW params",
		`defradb client index new --collection Users --fields vec `+
			`--vector '{"Metric":"COSINE","Dimensions":3,"HNSW":{"M":16,"EfConstruction":128,"EfSearch":64}}'`)

	cmd.Flags().StringVarP(&collectionArg, "collection", "c", "", "Collection name")
	cmd.Flags().StringVarP(&nameArg, "name", "n", "", "Index name")
	cmd.Flags().StringSliceVar(&fieldsArg, "fields", []string{}, "Fields to index")
	cmd.Flags().BoolVarP(&uniqueArg, "unique", "u", false, "Make the index unique")
	cmd.Flags().StringVar(&vectorArg, "vector", "", "Vector index config as JSON (makes a vector index)")

	return cmd
}
