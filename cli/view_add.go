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
	"strings"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"
	"github.com/spf13/cobra"
)

func MakeViewAddCommand() *cobra.Command {
	var query, sdl, lens string
	var queryFile, sdlFile, lensFile string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add new view",
		Long: `Add new database view.

Example: add from string flags:
  defradb client view add --query 'Foo { name, ...}' --sdl 'type Foo { ... }' --lens '{"lenses": [...'
Example: add from file flags:
  defradb client view add --query-file /path/to/query --sdl-file /path/to/sdl --lens-file /path/to/lens

Flag pairs <key>/<key>-file are mutually exclusive.

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetContextStore(cmd)

			query, err := pickDataOrReadFile(query, queryFile)
			if err != nil {
				return err
			}
			sdl, err := pickDataOrReadFile(sdl, sdlFile)
			if err != nil {
				return err
			}
			lensCfgJson, err := pickDataOrReadFile(lens, lensFile)
			if err != nil {
				return err
			}

			if lensCfgJson == "-" {
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				lensCfgJson = string(data)
			}

			var transform immutable.Option[model.Lens]
			if lensCfgJson != "" {
				decoder := json.NewDecoder(strings.NewReader(lensCfgJson))
				decoder.DisallowUnknownFields()

				var lensCfg model.Lens
				if err := decoder.Decode(&lensCfg); err != nil {
					return NewErrInvalidLensConfig(err)
				}
				transform = immutable.Some(lensCfg)
			}

			defs, err := store.AddView(cmd.Context(), query, sdl, transform)
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
	cmd.Flags().StringVarP(&lens, "lens", "", "", "Lens configuration")
	cmd.Flags().StringVarP(&lensFile, "lens-file", "", "", "Lens configuration file")

	cmd.MarkFlagsMutuallyExclusive("query", "query-file")
	cmd.MarkFlagsMutuallyExclusive("sdl", "sdl-file")
	cmd.MarkFlagsMutuallyExclusive("lens", "lens-file")
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
