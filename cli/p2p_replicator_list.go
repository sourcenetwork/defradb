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
)

func MakeP2PReplicatorListCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List all replicators",
		Long: `List all the replicators active in the P2P data sync system.
A replicator synchronizes one or all collection(s) from this instance to another.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			reps, err := cliClient.ListReplicators(cmd.Context())
			if err != nil {
				return err
			}
			return writeJSON(cmd, reps)
		},
	}

	EmbedCLIExample(ctx, cmd, "List all replicators",
		`defradb client p2p replicator list`)

	return cmd
}
