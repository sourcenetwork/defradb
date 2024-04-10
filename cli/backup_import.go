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

func MakeBackupImportCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "import <input_path>",
		Short: "Import a JSON data file to the database",
		Long: `Import a JSON data file to the database.

Example: import data to the database:
  defradb client import user_data.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			db := mustGetContextDB(cmd)
			return db.BasicImport(cmd.Context(), args[0])
		},
	}
	return cmd
}
