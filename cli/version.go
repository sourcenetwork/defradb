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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type versionInfo struct {
	Tag    string `json:"tag"`
	Commit string `json:"commit"`
	Date   string `json:"date"`
}

func (v versionInfo) FullVersion() string {
	return fmt.Sprintf(`DefraDB's Version Information:
  *  version tag  : %s
  *  build commit : %s
  *  release date : %s`,
		color.BlueString(DefraVersion.Tag),
		color.GreenString(DefraVersion.Commit),
		color.YellowString(DefraVersion.Date),
	)
}

func (v versionInfo) JsonVersion() (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

var DefraVersion = versionInfo{
	Tag:    "0.2.1",
	Commit: "e4328e0",
	Date:   "2022-03-07T00:12:07Z",
}

var format string

// versionCmd represents the command that will output the cli version
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version",
	RunE: func(cmd *cobra.Command, _ []string) error {
		switch strings.ToLower(format) {

		case "short":
			fmt.Println(DefraVersion.Tag) // nolint:forbidigo

		case "json":
			jVersion, err := DefraVersion.JsonVersion()
			if err != nil {
				return err
			}
			fmt.Println(jVersion) // nolint:forbidigo

		default:
			fmt.Println(DefraVersion.FullVersion()) // nolint:forbidigo

		}
		return nil
	},
}

func initVersionFormatFlag(cmd *cobra.Command) {
	fs := cmd.Flags()
	fs.SortFlags = false
	fs.StringVarP(&format, "format", "f", "", "The version's format can be one of: 'short', 'json'")
}

func init() {
	initVersionFormatFlag(versionCmd)
	rootCmd.AddCommand(versionCmd)
}
