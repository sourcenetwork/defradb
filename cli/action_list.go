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

	"github.com/sourcenetwork/defradb/client/options"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

func MakeActionListCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List action information.",
		Long: `List information pertaining to long running actions such as truncate, RefreshView,
and the (re)building of indexes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			opt := options.WithIdentity(options.ListActions(), iIdentity.FromContext(cmd.Context()))

			info, err := cliClient.ListActions(
				cmd.Context(),
				opt,
			)
			if err != nil {
				return err
			}

			return writeJSON(cmd, info)
		},
	}

	EmbedCLIExample(ctx, cmd, "List information about actions",
		"defradb client action list")

	return cmd
}
