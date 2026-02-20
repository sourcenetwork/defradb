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
	"context"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeSchemaAddCommand(ctx context.Context) *cobra.Command {
	var schemaFiles []string
	var cmd = &cobra.Command{
		Use:   "add [schema]",
		Short: "Add new schema",
		Long: `Add new schema.

Schema Object with a '@policy(id:".." resource: "..")' linked will only be accepted if:
  - ACP is available (i.e. ACP is not disabled).
  - The specified resource adheres to the document resource interface (DRI).
  - Learn more about the DefraDB [ACP System](https://docs.source.network/defradb/references/acp)

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := mustGetContextCLIClient(cmd)

			var combinedSchema string
			switch {
			case len(schemaFiles) > 0:
				// Read schemas from files and concatenate them
				for _, schemaFile := range schemaFiles {
					data, err := os.ReadFile(schemaFile)
					if err != nil {
						return NewErrFailedToReadSchemaFile(schemaFile, err)
					}
					combinedSchema += string(data) + "\n"
				}

			case len(args) > 0 && args[0] == "-":
				// Read schema from stdin
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return NewErrFailedToReadSchemaFromStdin(err)
				}
				combinedSchema += string(data) + "\n"

			case len(args) > 0:
				// Read schema from argument string
				combinedSchema += args[0] + "\n"

			default:
				return ErrEmptySchemaString
			}

			opt := options.WithIdentity(options.AddSchema(), identity.FromContext(cmd.Context()))
			// Process the combined schema
			cols, err := cli.AddSchema(cmd.Context(), combinedSchema, opt)
			if err != nil {
				return NewErrFailedToAddSchema(err)
			}
			if err := writeJSON(cmd, cols); err != nil {
				return err
			}

			return nil
		},
	}

	EmbedCLIExample(ctx, cmd, "add from an argument string",
		`defradb client schema add 'type Foo { ... }'`)

	EmbedCLIExample(ctx, cmd, "add from file",
		`defradb client schema add -f schema.graphql`)

	EmbedCLIExample(ctx, cmd, "add from multiple files",
		`defradb client schema add -f schema1.graphql -f schema2.graphql`)

	EmbedCLIExample(ctx, cmd, "add from multiple files (comma-separated)",
		`defradb client schema add -f schema1.graphql,schema2.graphql`)

	EmbedCLIExample(ctx, cmd, "add from stdin",
		`cat schema.graphql | defradb client schema add -`)

	cmd.Flags().StringSliceVarP(&schemaFiles, "file", "f", []string{}, "File to load schema from")
	return cmd
}
