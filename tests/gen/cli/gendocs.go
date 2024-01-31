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
	"io"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/tests/gen"
)

const defaultBatchSize = 1000

func MakeGenDocCommand() *cobra.Command {
	var demandJSON string

	var cmd = &cobra.Command{
		Use:   "gendocs --demand <demand_json>",
		Short: "Automatically generates documents for existing collections.",
		Long: `Automatically generates documents for existing collections.		

Example: The following command generates 100 User documents and 500 Device documents:
  gendocs --demand '{"User": 100, "Device": 500 }'`,
		ValidArgs: []string{"demand"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// cobra does not chain pre run calls so we have to run them again here
			if err := loadConfig(cfg); err != nil {
				return err
			}
			store, err := http.NewClient(cfg.API.Address)
			if err != nil {
				return err
			}

			demandMap := make(map[string]int)
			err = json.Unmarshal([]byte(demandJSON), &demandMap)
			if err != nil {
				return NewErrInvalidDemandValue(err)
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
			_, err = out.Write([]byte("Generated " + strconv.Itoa(len(docs)) +
				" documents. Adding to collections...\n"))
			if err != nil {
				return err
			}

			batchOffset := 0
			for batchOffset < len(docs) {
				batchLen := defaultBatchSize
				if batchOffset+batchLen > len(docs) {
					batchLen = len(docs) - batchOffset
				}

				colDocsMap := groupDocsByCollection(docs[batchOffset : batchOffset+batchLen])

				err = saveBatchToCollections(context.Background(), collections, colDocsMap)
				if err != nil {
					return err
				}

				err = reportSavedBatch(out, batchLen, colDocsMap)
				if err != nil {
					return err
				}

				batchOffset += batchLen
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&demandJSON, "demand", "d", "", "Documents' demand in JSON format")

	return cmd
}

func reportSavedBatch(out io.Writer, thisBatch int, colDocsMap map[string][]*client.Document) error {
	reports := make([]string, 0, len(colDocsMap))
	for colName, colDocs := range colDocsMap {
		reports = append(reports, strconv.Itoa(len(colDocs))+" "+colName)
	}

	r := strings.Join(reports, ", ")
	_, err := out.Write([]byte("Added " + strconv.Itoa(thisBatch) + " documents: " + r + "\n"))
	return err
}

func saveBatchToCollections(
	ctx context.Context,
	collections []client.Collection,
	colDocsMap map[string][]*client.Document,
) error {
	for colName, colDocs := range colDocsMap {
		for _, col := range collections {
			if col.Description().Name.Value() == colName {
				err := col.CreateMany(context.Background(), colDocs)
				if err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

func groupDocsByCollection(docs []gen.GeneratedDoc) map[string][]*client.Document {
	result := make(map[string][]*client.Document)
	for _, doc := range docs {
		result[doc.Col.Description.Name.Value()] = append(result[doc.Col.Description.Name.Value()], doc.Doc)
	}
	return result
}

func colsToDefs(cols []client.Collection) []client.CollectionDefinition {
	var colDefs []client.CollectionDefinition
	for _, col := range cols {
		colDefs = append(colDefs, col.Definition())
	}
	return colDefs
}

func loadConfig(cfg *config.Config) error {
	if err := cfg.LoadRootDirFromFlagOrDefault(); err != nil {
		return err
	}
	return cfg.LoadWithRootdir(cfg.ConfigFileExists())
}
