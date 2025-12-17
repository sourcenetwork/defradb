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

	"github.com/spf13/cobra"
)

func MakeLensListCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List all stored lenses",
		Long: `List all lenses stored in the lens store.

Returns a map of lens CIDs to their configurations.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			lenses, err := cliClient.ListLenses(cmd.Context())
			if err != nil {
				return err
			}

			return writeJSON(cmd, lenses)
		},
	}

	EmbedCLIExample(ctx, cmd, "list all lenses",
		`defradb client lens list`)

	return cmd
}
