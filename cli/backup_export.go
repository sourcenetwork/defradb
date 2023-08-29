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
	"strings"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/http"
)

const jsonFileType = "json"

func MakeBackupExportCommand(cfg *config.Config) *cobra.Command {
	var collections []string
	var pretty bool
	var format string
	var cmd = &cobra.Command{
		Use:   "export  [-c --collections | -p --pretty | -f --format] <output_path>",
		Short: "Export the database to a file",
		Long: `Export the database to a file. If a file exists at the <output_path> location, it will be overwritten.
		
If the --collection flag is provided, only the data for that collection will be exported.
Otherwise, all collections in the database will be exported.

If the --pretty flag is provided, the JSON will be pretty printed.

Example: export data for the 'Users' collection:
  defradb client export --collection Users user_data.json`,
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
			if !isValidExportFormat(format) {
				return ErrInvalidExportFormat
			}
			outputPath := args[0]

			for i := range collections {
				collections[i] = strings.Trim(collections[i], " ")
			}

			data := client.BackupConfig{
				Filepath:    outputPath,
				Format:      format,
				Pretty:      pretty,
				Collections: collections,
			}

			return db.BasicExport(cmd.Context(), &data)
		},
	}
	cmd.Flags().BoolVarP(&pretty, "pretty", "p", false, "Set the output JSON to be pretty printed")
	cmd.Flags().StringVarP(&format, "format", "f", jsonFileType,
		"Define the output format. Supported formats: [json]")
	cmd.Flags().StringSliceVarP(&collections, "collections", "c", []string{}, "List of collections")

	return cmd
}

func isValidExportFormat(format string) bool {
	switch strings.ToLower(format) {
	case jsonFileType:
		return true
	default:
		return false
	}
}
