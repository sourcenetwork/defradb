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
	"context"

	"github.com/spf13/cobra"
)

func MakeKeyringGetCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get <name>",
		Short: "Get a private key",
		Long: `Get a private key.
Prints the hexadecimal representation of a private key.

The DEFRA_KEYRING_SECRET environment variable must be set to unlock the keyring.
This can also be done with a .env file in the working directory or at a path
defined with the --secret-file flag.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyring, err := openKeyring(cmd)
			if err != nil {
				return err
			}
			keyBytes, err := keyring.Get(args[0])
			if err != nil {
				return err
			}
			cmd.Printf("%x\n", keyBytes)
			return nil
		},
	}

	EmbedCLIExample(ctx, cmd, "Get encryption key",
		`defradb keyring get encryption-key`)

	return cmd
}
