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
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/tests/gen"
)

const gendocBatchSize = 1000

func MakeGenDocCommand(cfg *config.Config) *cobra.Command {
	var demandJSON string

	var cmd = &cobra.Command{
		Use:   "gendocs --demand <demand_json>",
		Short: "Automatically generates documents for existing collections.",
		Long: `Automatically generates documents for existing collections.		

Example: generates 100 User documents and 500 Device documents:
  defradb gendocs --demand '{"User": 100, "Device": 500 }'`,
		ValidArgs: []string{"demand"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// cobra does not chain pre run calls so we have to run them again here
			if err := loadConfig(cfg); err != nil {
				return err
			}
			if err := setTransactionContext(cmd, cfg, 0); err != nil {
				return err
			}
			if err := setStoreContext(cmd, cfg); err != nil {
				return err
			}
			store := mustGetStoreContext(cmd)

			demandMap := make(map[string]int)
			err := json.Unmarshal([]byte(demandJSON), &demandMap)
			if err != nil {
				return err
			}

			collections, err := store.GetAllCollections(cmd.Context())
			if err != nil {
				return err
			}

			opts := []gen.Option{}
			for colName, numDocs := range demandMap {
				opts = append(opts, gen.WithTypeDemand(colName, numDocs))
			}
			docs, err := gen.AutoGenerate(colsToDefs(collections), opts...)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			_, err = out.Write([]byte("Generated " + strconv.Itoa(len(docs)) + " documents\n"))
			if err != nil {
				return err
			}

			for len(docs) > 0 {
				thisBatch := gendocBatchSize
				if len(docs) < gendocBatchSize {
					thisBatch = len(docs)
				}

				colDocsMap := make(map[string][]*client.Document)
				for _, doc := range docs[:thisBatch] {
					colDocsMap[doc.Col.Description.Name] = append(colDocsMap[doc.Col.Description.Name], doc.Doc)
				}

				for colName, colDocs := range colDocsMap {
					for i := range collections {
						if collections[i].Description().Name == colName {
							err = collections[i].CreateMany(context.Background(), colDocs)
							if err != nil {
								return err
							}
							break
						}
					}
				}

				docs = docs[thisBatch:]

				reports := make([]string, 0, len(colDocsMap))
				for colName, colDocs := range colDocsMap {
					reports = append(reports, strconv.Itoa(len(colDocs))+" "+colName)
				}

				r := strings.Join(reports, ", ")
				_, err = out.Write([]byte("Added " + strconv.Itoa(thisBatch) + " documents: " + r + "\n"))
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&demandJSON, "demand", "d", "", "Documents' demand in JSON format")

	return cmd
}

func colsToDefs(cols []client.Collection) []client.CollectionDefinition {
	var colDefs []client.CollectionDefinition
	for _, col := range cols {
		colDefs = append(colDefs, col.Definition())
	}
	return colDefs
}
