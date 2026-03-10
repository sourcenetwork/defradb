// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package cli

import (
	"context"
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/tests/gen"
)

const defaultBatchSize = 1000

func MakeGenDocCommand(ctx context.Context) *cobra.Command {
	var demandJSON string
	var url string
	var cmd = &cobra.Command{
		Use:   "gendocs --demand <demand_json>",
		Short: "Automatically generates documents for existing collections.",
		Long: `Automatically generates documents for existing collections.		

Example: The following command generates 100 User documents and 500 Device documents:
  gendocs --demand '{"User": 100, "Device": 500 }'`,
		ValidArgs: []string{"demand"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := http.NewClient(url)
			if err != nil {
				return err
			}

			demandMap := make(map[string]int)
			err = json.Unmarshal([]byte(demandJSON), &demandMap)
			if err != nil {
				return NewErrInvalidDemandValue(err)
			}

			collections, err := c.GetCollections(cmd.Context())
			if err != nil {
				return err
			}

			opts := []gen.Option{}
			for colName, numDocs := range demandMap {
				opts = append(opts, gen.WithTypeDemand(colName, numDocs))
			}
			docs, err := gen.AutoGenerate(ctx, colsToVersions(collections), opts...)
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

	cmd.Flags().StringVar(&url, "url", "localhost:9181", "URL of HTTP endpoint to listen on or connect to")
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
			if col.Version().Name == colName {
				err := col.AddManyDocuments(ctx, colDocs)
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
		result[doc.Col.Name] = append(result[doc.Col.Name], doc.Doc)
	}
	return result
}

func colsToVersions(cols []client.Collection) []client.CollectionVersion {
	var colDefs []client.CollectionVersion
	for _, col := range cols {
		colDefs = append(colDefs, col.Version())
	}
	return colDefs
}
