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

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/http"
)

func MakeBackupImportCommand(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "import <input_path>",
		Short: "Import a JSON data file to the database",
		Long: `Import a JSON data file to the database.

Example: import data to the database:
  defradb client import user_data.json`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return NewErrInvalidArgumentLength(err, 1)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			db, err := http.NewClient("http://" + cfg.API.Address)
			if err != nil {
				return err
			}
			return db.BasicImport(cmd.Context(), args[0])
		},
	}
	return cmd
}
