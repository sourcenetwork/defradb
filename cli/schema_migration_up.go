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
	"io"
	"os"

	"github.com/sourcenetwork/immutable/enumerable"
	"github.com/spf13/cobra"
)

func MakeSchemaMigrationUpCommand() *cobra.Command {
	var file string
	var collectionID uint32
	var cmd = &cobra.Command{
		Use:   "up --collection <collectionID> <documents>",
		Short: "Applies the migration to the specified collection version.",
		Long: `Applies the migration to the specified collection version.
Documents is a list of documents to apply the migration to.		

Example: migrate from string
  defradb client schema migration up --collection 2 '[{"name": "Bob"}]'

Example: migrate from file
  defradb client schema migration up --collection 2 -f documents.json

Example: migrate from stdin
  cat documents.json | defradb client schema migration up --collection 2 -
		`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db := mustGetContextDB(cmd)

			var srcData []byte
			switch {
			case file != "":
				data, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				srcData = data
			case len(args) == 1 && args[0] == "-":
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				srcData = data
			case len(args) == 1:
				srcData = []byte(args[0])
			default:
				return ErrNoDocOrFile
			}

			var src []map[string]any
			if err := json.Unmarshal(srcData, &src); err != nil {
				return err
			}
			out, err := db.LensRegistry().MigrateUp(cmd.Context(), enumerable.New(src), collectionID)
			if err != nil {
				return err
			}
			var value []map[string]any
			err = enumerable.ForEach(out, func(item map[string]any) {
				value = append(value, item)
			})
			if err != nil {
				return err
			}
			return writeJSON(cmd, value)
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "File containing document(s)")
	cmd.Flags().Uint32Var(&collectionID, "collection", 0, "Collection id")
	return cmd
}
