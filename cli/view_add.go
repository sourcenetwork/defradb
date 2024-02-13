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

import "github.com/spf13/cobra"

func MakeViewAddCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add [query] [sdl]",
		Short: "Add new view",
		Long: `Add new database view.

Example: add from an argument string:
  defradb client view add 'Foo { name, ...}' 'type Foo { ... }'

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetContextStore(cmd)

			if len(args) != 2 {
				return ErrViewAddMissingArgs
			}

			query := args[0]
			sdl := args[1]

			defs, err := store.AddView(cmd.Context(), query, sdl)
			if err != nil {
				return err
			}
			return writeJSON(cmd, defs)
		},
	}
	return cmd
}
