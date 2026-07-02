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

func MakeKVImportCommand(ctx context.Context) *cobra.Command {
	var rebuildIndexes string

	var cmd = &cobra.Command{
		Use:   "import <input_path>",
		Short: "Import raw KV pairs from a file into the database",
		Long: `Import raw KV pairs from a binary file into the database's rootstore.

The input file must use the DefraDB KV wire format produced by 'kv export'.

Use --rebuild-indexes to rebuild secondary indexes for a collection after import.

NOTE: This command requires direct node access and is not available via the HTTP client.
Use it against a local node instance (not a remote HTTP endpoint).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath := args[0]

			f, err := os.Open(inputPath)
			if err != nil {
				return err
			}
			defer f.Close()

			cliClient := mustGetContextCLIClient(cmd)
			n, err := cliClient.ImportRawKVs(cmd.Context(), f)
			if err != nil {
				return err
			}
			cmd.Printf("Imported %d KV entries from %s\n", n, inputPath)

			if rebuildIndexes != "" {
				if err := cliClient.RebuildCollectionIndexes(cmd.Context(), rebuildIndexes); err != nil {
					return err
				}
				cmd.Printf("Rebuilt indexes for collection %q\n", rebuildIndexes)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&rebuildIndexes, "rebuild-indexes", "",
		"Collection name whose indexes should be rebuilt after import")

	return cmd
}
