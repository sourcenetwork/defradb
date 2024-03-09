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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"
	"github.com/spf13/cobra"
)

func MakeSchemaPatchCommand() *cobra.Command {
	var patchFile string
	var lensFile string
	var setActive bool
	var cmd = &cobra.Command{
		Use:   "patch [schema] [migration]",
		Short: "Patch an existing schema type",
		Long: `Patch an existing schema.

Uses JSON Patch to modify schema types.

Example: patch from an argument string:
  defradb client schema patch '[{ "op": "add", "path": "...", "value": {...} }]' '{"lenses": [...'

Example: patch from file:
  defradb client schema patch -f patch.json

Example: patch from stdin:
  cat patch.json | defradb client schema patch -

To learn more about the DefraDB GraphQL Schema Language, refer to https://docs.source.network.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetContextStore(cmd)

			var patch string
			switch {
			case patchFile != "":
				data, err := os.ReadFile(patchFile)
				if err != nil {
					return err
				}
				patch = string(data)
			case len(args) > 0 && args[0] == "-":
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				patch = string(data)
			case len(args) >= 1:
				patch = args[0]
			default:
				return fmt.Errorf("patch cannot be empty")
			}

			var lensCfgJson string
			switch {
			case lensFile != "":
				data, err := os.ReadFile(lensFile)
				if err != nil {
					return err
				}
				patch = string(data)
			case len(args) == 2:
				lensCfgJson = args[1]
			}

			decoder := json.NewDecoder(strings.NewReader(lensCfgJson))
			decoder.DisallowUnknownFields()

			var migration immutable.Option[model.Lens]
			if lensCfgJson != "" {
				var lensCfg model.Lens
				if err := decoder.Decode(&lensCfg); err != nil {
					return NewErrInvalidLensConfig(err)
				}
				migration = immutable.Some(lensCfg)
			}

			return store.PatchSchema(cmd.Context(), patch, migration, setActive)
		},
	}
	cmd.Flags().BoolVar(&setActive, "set-active", false,
		"Set the active schema version for all collections using the root schem")
	cmd.Flags().StringVarP(&patchFile, "patch-file", "p", "", "File to load a patch from")
	cmd.Flags().StringVarP(&lensFile, "lens-file", "t", "", "File to load a lens config from")
	return cmd
}
