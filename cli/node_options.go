// Copyright 2026 Democratized Data Foundation
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

func MakeNodeOptionsCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "node-options",
		Short: "Get the node's configuration options as JSON",
		Long:  `Get the node's configuration options as JSON.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)
			opts, err := cliClient.GetNodeOptions(cmd.Context())
			if err != nil {
				return err
			}
			return writeJSON(cmd, opts)
		},
	}
	return cmd
}
