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

	"github.com/sourcenetwork/defradb/client"
)

func MakeSchemaSetDefaultCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "set-default [versionID]",
		Short: "Set the default schema version",
		Long:  `Set the default schema version`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := cmd.Context().Value(storeContextKey).(client.Store)
			return store.SetDefaultSchemaVersion(cmd.Context(), args[0])
		},
	}
	return cmd
}
