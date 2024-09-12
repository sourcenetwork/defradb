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
	"os"

	"github.com/spf13/cobra"
)

func MakePurgeCommand() *cobra.Command {
	var force bool
	var cmd = &cobra.Command{
		Use:   "purge",
		Short: "Delete all persisted DefraDB data",
		Long: `Delete all persisted DefraDB data.
WARNING this operation will delete all data and cannot be reversed.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := setContextRootDir(cmd); err != nil {
				return err
			}
			return setContextConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				return ErrPurgeForceFlagRequired
			}
			cfg := mustGetContextConfig(cmd)
			return os.RemoveAll(cfg.GetString("datastore.badger.path"))
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Must be set for the operation to run")
	return cmd
}
