// Copyright 2025 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/wizard"
)

// Exports the createConfig function for use by the setup wizard
func createDefaultConfig(rootdir string) error {
	defaultCmd := &cobra.Command{}
	return createConfig(rootdir, defaultCmd.Flags())
}

func MakeWizardCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wizard",
		Short: "Runs the DefraDB setup wizard",
		RunE: func(cmd *cobra.Command, _ []string) error {
			wizard.Main(createDefaultConfig)
			return nil
		},
	}
	return cmd
}
