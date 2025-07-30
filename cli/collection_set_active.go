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

func MakeCollectionSetActiveCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "set-active [versionID]",
		Short: "Set the active collection version",
		Long: `Activates all collection versions with the given version id, and deactivates all
other versions of that collection.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)
			return cliClient.SetActiveCollectionVersion(cmd.Context(), args[0])
		},
	}
	return cmd
}
