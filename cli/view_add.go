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
	var lensFile string
	var cmd = &cobra.Command{
		Use:   "add [query] [sdl] [transform]",
		Short: "Add new view",
		Long: `Add new database view.

Example: add from an argument string:
  defradb client view add 'Foo { name, ...}' 'type Foo { ... }' '{"lenses": [...'

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.`,
		Args: cobra.RangeArgs(2, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetStoreContext(cmd)

			query := args[0]
			sdl := args[1]

			var lensCfgJson string
			switch {
			case lensFile != "":
				data, err := os.ReadFile(lensFile)
				if err != nil {
					return err
				}
				lensCfgJson = string(data)
			case len(args) == 3 && args[2] == "-":
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				lensCfgJson = string(data)
			case len(args) == 3:
				lensCfgJson = args[2]
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
	cmd.Flags().StringVarP(&lensFile, "file", "f", "", "Lens configuration file")
	return cmd
}
