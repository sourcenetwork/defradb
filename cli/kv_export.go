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
	"os"

	"github.com/spf13/cobra"
)

func MakeKVExportCommand(ctx context.Context) *cobra.Command {
	var docIDs []string
	var datastoreOnly bool

	var cmd = &cobra.Command{
		Use:   "export <collection> <output_path>",
		Short: "Export raw KV pairs for specified documents to a file",
		Long: `Export raw KV pairs for the specified documents to a binary file.

The output file uses the DefraDB KV wire format: for each entry, a uint32 namespace byte
followed by the raw key and value. A zero-length sentinel marks the end of the stream.

Use --doc-ids to restrict the export to specific document IDs.
Use --datastore-only to skip headstore and blockstore entries.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			collectionName := args[0]
			outputPath := args[1]

			f, err := os.Create(outputPath)
			if err != nil {
				return err
			}
			defer f.Close()

			cliClient := mustGetContextCLIClient(cmd)
			n, err := cliClient.ExportDocKVs(cmd.Context(), collectionName, docIDs, f, datastoreOnly)
			if err != nil {
				return err
			}
			cmd.Printf("Exported %d KV entries to %s\n", n, outputPath)
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&docIDs, "doc-ids", nil, "Document IDs to export (exports all if omitted)")
	cmd.Flags().BoolVar(&datastoreOnly, "datastore-only", false, "Skip headstore and blockstore entries")

	return cmd
}
