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

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/immutable"
)

func MakeViewAddCommand(ctx context.Context) *cobra.Command {
	var lensCID string
	var cmd = &cobra.Command{
		Use:   "add [query] [sdl]",
		Short: "Add new view",
		Long: `Add new database view.

Use --lens-cid to specify a lens transform. Store a lens first using 'defradb client lens add'.

Learn more about the DefraDB GraphQL Schema Language on https://docs.source.network.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			query := args[0]
			sdl := args[1]

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

	EmbedCLIExample(ctx, cmd, "add a simple view",
		`defradb client view add 'Foo { name, ...}' 'type Foo { ... }'`)
	EmbedCLIExample(ctx, cmd, "add using an existing lens CID",
		`defradb client view add 'Foo { name, ...}' 'type Foo { ... }' --lens-cid bafyreih...`)

	cmd.Flags().StringVar(&lensCID, "lens-cid", "", "CID of an existing lens transform (use 'lens add' first)")
	return cmd
}
