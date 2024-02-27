// Copyright 2022 Democratized Data Foundation
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
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func MakeSchemaAddCommand() *cobra.Command {
	var schemaFile string
	var cmd = &cobra.Command{
		Use:   "add [schema]",
		Short: "Add new schema",
		Long: `Add new schema.

Schema Object with a '@policy(id:".." resource: "..")' linked will only be accepted if:
  - ACP module is available (i.e. ACP is not disabled).
  - The specified resource adheres to the Document Access Control DPI Rules.
  - Learn more about [ACP & DPI Rules](/acp/README.md)

Example: add from an argument string:
  defradb client schema add 'type Foo { ... }'

Example: add from file:
  defradb client schema add -f schema.graphql

Example: add from stdin:
  cat schema.graphql | defradb client schema add -

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetContextStore(cmd)

			var schema string
			switch {
			case schemaFile != "":
				data, err := os.ReadFile(schemaFile)
				if err != nil {
					return err
				}
				schema = string(data)
			case len(args) > 0 && args[0] == "-":
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				schema = string(data)
			case len(args) > 0:
				schema = args[0]
			default:
				return fmt.Errorf("schema cannot be empty")
			}

			cols, err := store.AddSchema(cmd.Context(), schema)
			if err != nil {
				return err
			}
			return writeJSON(cmd, cols)
		},
	}
	cmd.Flags().StringVarP(&schemaFile, "file", "f", "", "File to load a schema from")
	return cmd
}
