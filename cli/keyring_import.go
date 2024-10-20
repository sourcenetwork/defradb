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
	"encoding/hex"

	"github.com/spf13/cobra"
)

func MakeKeyringImportCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "import <name> <private-key-hex>",
		Short: "Import a private key",
		Long: `Import a private key.
Store an externally generated key in the keyring.

The DEFRA_KEYRING_SECRET environment variable must be set to unlock the keyring.
This can also be done with a .env file in the working directory or at a path
defined with the --secret-file flag.

Example:
  defradb keyring import encryption-key 0000000000000000`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyring, err := openKeyring(cmd)
			if err != nil {
				return err
			}
			keyBytes, err := hex.DecodeString(args[1])
			if err != nil {
				return err
			}
			return keyring.Set(args[0], keyBytes)
		},
	}
	return cmd
}
