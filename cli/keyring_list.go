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
)

// MakeKeyringListCommand creates a new command to list all keys in the keyring.
func MakeKeyringListCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List all keys in the keyring",
		Long: `List all keys in the keyring.
The DEFRA_KEYRING_SECRET environment variable must be set to unlock the keyring.
This can also be done with a .env file in the working directory or at a path
defined with the --secret-file flag.

Example:
  defradb keyring list`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			keyring, err := openKeyring(cmd)
			if err != nil {
				return err
			}

			keyNames, err := keyring.List()
			if err != nil {
				return err
			}

			if len(keyNames) == 0 {
				cmd.Println("No keys found in the keyring.")
				return nil
			}

			cmd.Println("Keys in the keyring:")
			for _, keyName := range keyNames {
				cmd.Println("- " + keyName)
			}
			return nil
		},
	}
	return cmd
}
