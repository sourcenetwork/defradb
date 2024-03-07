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
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func MakeCollectionPatchCommand() *cobra.Command {
	var patchFile string
	var cmd = &cobra.Command{
		Use:   "patch [patch]",
		Short: "Patch existing collection descriptions",
		Long: `Patch existing collection descriptions.

Uses JSON Patch to modify collection descriptions.

Example: patch from an argument string:
  defradb client collection patch '[{ "op": "add", "path": "...", "value": {...} }]'

Example: patch from file:
  defradb client collection patch -f patch.json

Example: patch from stdin:
  cat patch.json | defradb client collection patch -

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

			return store.PatchCollection(cmd.Context(), patch)
		},
	}
	cmd.Flags().StringVarP(&patchFile, "patch-file", "p", "", "File to load a patch from")
	return cmd
}
