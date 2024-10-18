// Copyright 2024 Democratized Data Foundation
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

func MakePurgeCommand() *cobra.Command {
	var force bool
	var cmd = &cobra.Command{
		Use:   "purge",
		Short: "Delete all persisted data and restart",
		Long: `Delete all persisted data and restart.
WARNING this operation cannot be reversed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			db := mustGetContextHTTP(cmd)
			if !force {
				return ErrPurgeForceFlagRequired
			}
			return db.Purge(cmd.Context())
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Must be set for the operation to run")
	return cmd
}
