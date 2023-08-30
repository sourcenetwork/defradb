// Copyright 2022 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/version"
)

func MakeVersionCommand() *cobra.Command {
	var format string
	var full bool
	var cmd = &cobra.Command{
		Use:   "version",
		Short: "Display the version information of DefraDB and its components",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dv, err := version.NewDefraVersion()
			if err != nil {
				return err
			}

			if strings.ToLower(format) == "json" {
				return writeJSON(cmd, dv)
			}

			if full {
				cmd.Println(dv.StringFull())
			} else {
				cmd.Println(dv.String())
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&format, "format", "f", "", "Version output format. Options are text, json")
	cmd.Flags().BoolVarP(&full, "full", "", false, "Display the full version information")
	return cmd
}
