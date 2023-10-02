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
	"github.com/spf13/cobra"
)

func MakeP2PReplicatorGetAllCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "getall",
		Short: "Get all replicators",
		Long: `Get all the replicators active in the P2P data sync system.
These are the replicators that are currently replicating data from one node to another.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetStoreContext(cmd)

			reps, err := store.GetAllReplicators(cmd.Context())
			if err != nil {
				return err
			}
			return writeJSON(cmd, reps)
		},
	}
	return cmd
}
