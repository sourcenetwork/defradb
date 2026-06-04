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

func MakeCollectionAddCommand(ctx context.Context) *cobra.Command {
	var sdlFiles []string
	var cmd = &cobra.Command{
		Use:   "add [sdl]",
		Short: "Add new collection",
		Long: `Add new collection.

Collection type with a '@policy(id:".." resource: "..")' linked will only be accepted if:
  - ACP is available (i.e. ACP is not disabled).
  - The specified resource adheres to the document resource interface (DRI).
  - Learn more about the DefraDB [ACP System](https://docs.source.network/defradb/references/acp)

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := mustGetContextCLIClient(cmd)

			var combinedSDL string
			switch {
			case len(sdlFiles) > 0:
				// Read collection definitions from files and concatenate them
				for _, sdlFile := range sdlFiles {
					data, err := os.ReadFile(sdlFile)
					if err != nil {
						return NewErrFailedToReadCollectionFile(sdlFile, err)
					}
					combinedSDL += string(data) + "\n"
				}

			case len(args) > 0 && args[0] == "-":
				// Read collection definition from stdin
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return NewErrFailedToReadCollectionFromStdin(err)
				}
				combinedSDL += string(data) + "\n"

			case len(args) > 0:
				// Read collection definition from argument string
				combinedSDL += args[0] + "\n"

			default:
				return ErrEmptyCollectionSDL
			}

			opt := options.WithIdentity(options.AddCollection(), identity.FromContext(cmd.Context()))
			// Process the combined SDL
			cols, err := cli.AddCollection(cmd.Context(), combinedSDL, opt)
			if err != nil {
				return NewErrFailedToAddCollection(err)
			}
			if err := writeJSON(cmd, cols); err != nil {
				return err
			}

			return nil
		},
	}

	EmbedCLIExample(ctx, cmd, "add from an argument string",
		`defradb client collection add 'type Foo { ... }'`)

	EmbedCLIExample(ctx, cmd, "add from file",
		`defradb client collection add -f schema.graphql`)

	EmbedCLIExample(ctx, cmd, "add from multiple files",
		`defradb client collection add -f schema1.graphql -f schema2.graphql`)

	EmbedCLIExample(ctx, cmd, "add from multiple files (comma-separated)",
		`defradb client collection add -f schema1.graphql,schema2.graphql`)

	EmbedCLIExample(ctx, cmd, "add from stdin",
		`cat schema.graphql | defradb client collection add -`)

	cmd.Flags().StringSliceVarP(&sdlFiles, "file", "f", []string{}, "File to load a collection definition from")
	return cmd
}
