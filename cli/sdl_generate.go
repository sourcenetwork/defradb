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
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema"
)

var (
	defaultOutputPath = "schema.gen.graphql"
	fileLineSeperator = "\n\n"
)

func MakeSDLGenerateCommand(ctx context.Context) *cobra.Command {
	var outputFile string
	var yesOverwrite bool
	var searchableEncryption bool
	var cmd = &cobra.Command{
		Use:   "generate --output schema.graphql <input schema files...>",
		Short: "Generate full GraphQL formatted schema.",
		Long: `Generates the fully formatted GraphQL schema from a given user type definition(s).

Accepts multiple input files as well as "-" to use stdin.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var sdlBuf string

			// Either we use stdin or we concat all the file
			// arguments
			if len(args) == 1 && args[0] == "-" {
				sdlByteBuf, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return NewErrReadingArgument("stdin", err)
				}
				sdlBuf = string(sdlByteBuf)
			} else {
				var fileInputBuf strings.Builder
				for i, arg := range args {
					if arg == "-" {
						return ErrStdinSingleInputOnly
					}
					fileBuf, err := os.ReadFile(arg)
					if err != nil {
						return NewErrReadingArgument("file", err)
					}

					if i != 0 {
						fileInputBuf.WriteString(fileLineSeperator)
					}
					fileInputBuf.Write(fileBuf)
				}
				sdlBuf = fileInputBuf.String()
			}

			var outWriter io.Writer
			if outputFile == "-" {
				outWriter = cmd.OutOrStdout()
			} else {
				// check if the file exists, if so check for the overwrite
				// flag
				ofinfo, err := os.Stat(outputFile)
				if err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}
				if ofinfo != nil && !yesOverwrite {
					return errors.New("output file path already exists. If you want to overwrite use -y")
				}

				f, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					return err
				}
				defer f.Close() //nolint:errcheck
				outWriter = f
			}

			schemaManager, err := schema.NewSchemaManager(searchableEncryption)
			if err != nil {
				return err
			}

			cols, err := schemaManager.ParseSDL(sdlBuf)
			if err != nil {
				return errors.Join(ErrParsingSDL, err)
			}

			collections := make([]client.CollectionVersion, len(cols))
			for i, c := range cols {
				collections[i] = c.Definition
			}

			_, err = schemaManager.Generator.Generate(ctx, collections)
			if err != nil {
				return errors.Join(ErrGeneratingSDL, err)
			}

			return schemaManager.WriteSDL(outWriter)
		},
	}

	EmbedCLIExample(ctx, cmd, "Generate SDL",
		`defradb sdl generate foo.graphql`)

	EmbedCLIExample(ctx, cmd, "Generate Multiple SDLs",
		`defradb sdl generate foo.graphql bar.graphql`)

	EmbedCLIExample(ctx, cmd, "Generate SDL and overwrite output",
		`defradb sdl generate foo.graphql bar.graphql --output schema.graphql -y`)

	cmd.PersistentFlags().StringVarP(&outputFile, "output", "o", defaultOutputPath,
		"The output file to write the generated schema. Accepts '-' to write to stdout")

	EmbedCLIExample(ctx, cmd, "Generate SDL with Searchable Encryption type definitions",
		`defradb sdl generate foo.graphql -s`)

	cmd.PersistentFlags().BoolVarP(&yesOverwrite, "overwrite", "y", false,
		"Overwrite any existing matching output file paths")

	cmd.PersistentFlags().BoolVarP(&searchableEncryption, "include-searchable-encryption", "s",
		false, "Include the schema type definitions to support Searchable Encryption")

	return cmd
}
