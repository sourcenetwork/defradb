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

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/http"
)

func MakeSchemaPatchCommand(cfg *config.Config) *cobra.Command {
	var patchFile string

	var cmd = &cobra.Command{
		Use:   "patch [schema]",
		Short: "Patch an existing schema type",
		Long: `Patch an existing schema.

Uses JSON Patch to modify schema types.

Example: patch from an argument string:
  defradb client schema patch '[{ "op": "add", "path": "...", "value": {...} }]'

Example: patch from file:
  defradb client schema patch -f patch.json

Example: patch from stdin:
  cat patch.json | defradb client schema patch -

To learn more about the DefraDB GraphQL Schema Language, refer to https://docs.source.network.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := http.NewClient("http://" + cfg.API.Address)
			if err != nil {
				return err
			}

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
			case len(args) > 0:
				patch = args[0]
			default:
				return fmt.Errorf("patch cannot be empty")
			}

			return db.PatchSchema(cmd.Context(), patch)
		},
	}
	cmd.Flags().StringVarP(&patchFile, "file", "f", "", "File to load a patch from")
	return cmd
}
