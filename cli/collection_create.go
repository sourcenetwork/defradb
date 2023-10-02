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

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
)

func MakeCollectionCreateCommand() *cobra.Command {
	var file string
	var cmd = &cobra.Command{
		Use:   "create <document>",
		Short: "Create a new document.",
		Long: `Create a new document.

Example: create from string
  defradb client collection create --name User '{ "name": "Bob" }'

Example: create multiple from string
  defradb client collection create --name User '[{ "name": "Alice" }, { "name": "Bob" }]'

Example: create from file
  defradb client collection create --name User -f document.json

Example: create from stdin
  cat document.json | defradb client collection create --name User -
		`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			col, ok := tryGetCollectionContext(cmd)
			if !ok {
				return cmd.Usage()
			}

			var docData []byte
			switch {
			case file != "":
				data, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				docData = data
			case len(args) == 1 && args[0] == "-":
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				docData = data
			case len(args) == 1:
				docData = []byte(args[0])
			default:
				return ErrNoDocOrFile
			}

			var docMap any
			if err := json.Unmarshal(docData, &docMap); err != nil {
				return err
			}

			switch t := docMap.(type) {
			case map[string]any:
				doc, err := client.NewDocFromMap(t)
				if err != nil {
					return err
				}
				return col.Create(cmd.Context(), doc)
			case []any:
				docs := make([]*client.Document, len(t))
				for i, v := range t {
					docMap, ok := v.(map[string]any)
					if !ok {
						return ErrInvalidDocument
					}
					doc, err := client.NewDocFromMap(docMap)
					if err != nil {
						return err
					}
					docs[i] = doc
				}
				return col.CreateMany(cmd.Context(), docs)
			default:
				return ErrInvalidDocument
			}
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "File containing document(s)")
	return cmd
}
