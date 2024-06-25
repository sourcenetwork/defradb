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
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/acp/identity"
)

func MakeIdentityNewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "new",
		Short: "Generate a new identity",
		Long: `Generate a new identity

Example: generate a new identity:
  defradb identity new

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			newIdentity, err := identity.Generate()
			if err != nil {
				return err
			}

			return writeJSON(cmd, newIdentity)
		},
	}

	return cmd
}
