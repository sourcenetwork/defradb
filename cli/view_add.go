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
	"os"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/immutable"
)

func MakeViewAddCommand(ctx context.Context) *cobra.Command {
	var query, sdl, lensCID string
	var queryFile, sdlFile string
	cmd := &cobra.Command{
		Use:   "add [query|query-file] [sdl|sdl-file]",
		Short: "Add new view",
		Long: `Add new database view.

Use --lens-cid to specify a lens transform. Store a lens first using 'defradb client lens add

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			query, err := pickDataOrReadFile(query, queryFile)
			if err != nil {
				return err
			}
			sdl, err := pickDataOrReadFile(sdl, sdlFile)
			if err != nil {
				return err
			}
			var transformCIDOpt immutable.Option[string]
			if lensCID != "" {
				transformCIDOpt = immutable.Some(lensCID)
			}

			defs, err := cliClient.AddView(cmd.Context(), query, sdl, transformCIDOpt)
			if err != nil {
				return err
			}
			return writeJSON(cmd, defs)
		},
	}
	cmd.Flags().StringVarP(&query, "query", "", "", "Query")
	cmd.Flags().StringVarP(&queryFile, "query-file", "", "", "Query file")
	cmd.Flags().StringVarP(&sdl, "sdl", "", "", "SDL")
	cmd.Flags().StringVarP(&sdlFile, "sdl-file", "", "", "SDL file")
	cmd.Flags().StringVar(&lensCID, "lens-cid", "", "CID of an existing lens transform (use 'lens add' first)")

	cmd.MarkFlagsMutuallyExclusive("query", "query-file")
	cmd.MarkFlagsMutuallyExclusive("sdl", "sdl-file")

	EmbedCLIExample(ctx, cmd, "add a simple view from string flags",
		`defradb client view add --query 'Foo { name, ...}' --sdl 'type Foo { ... }'`)
	EmbedCLIExample(ctx, cmd, "add using an existing lens CID",
		`defradb client view add --query-file /path/to/query --sdl-file /path/to/sdl --lens-cid bafyreih...`)
	EmbedCLIExample(ctx, cmd, "add from file flags using an existing lens CID",
		`defradb client view add --query-file /path/to/query --sdl-file /path/to/sdl --lens-cid bafyreih...`)

	return cmd
}

// pickDataOrReadFile gets the result from file path when provided, or from data.
func pickDataOrReadFile(data string, dataPath string) (string, error) {
	if dataPath == "" {
		return data, nil
	}
	b, err := os.ReadFile(dataPath)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
