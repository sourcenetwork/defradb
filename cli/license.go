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
	"github.com/sourcenetwork/defradb/licenses"
	"github.com/spf13/cobra"
)

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Display license information",
	Long: `Display the license information of DefraDB, which uses a parametric license model called
the Business Source License (BSL).`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(licenses.LicenseBSL)
	},
}

func init() {
	rootCmd.AddCommand(licenseCmd)
}
