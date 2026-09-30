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

	"github.com/sourcenetwork/defradb/cli/config"
)

func MakeKeyringCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "keyring",
		Short: "Manage DefraDB private keys",
		Long: `Manage DefraDB private keys.
Generate, add, get, and list private keys.

The following keys are loaded from the keyring on start:

- peer-key: Ed25519 key (required)
- encryption-key: AES-128, AES-192, or AES-256 key (optional)
`,
	}

	EmbedCLIExample(ctx, cmd, "Randomly generate the required keys",
		"defradb keyring new")
	EmbedCLIExample(ctx, cmd, "Import externally generated keys",
		"defradb keyring add <name> <private-key-hex>")

	setKeyringFlags(cmd)

	return cmd
}

func setKeyringFlags(cmd *cobra.Command) {
	cfg := config.DefaultConfig()
	cmd.PersistentFlags().String(
		"keyring-namespace",
		cfg.GetString(config.ConfigFlags["keyring-namespace"]),
		"Service name to use when using the system backend",
	)
	cmd.PersistentFlags().String(
		"keyring-backend",
		cfg.GetString(config.ConfigFlags["keyring-backend"]),
		"Keyring backend to use. Options are file or system",
	)
	cmd.PersistentFlags().String(
		"keyring-path",
		cfg.GetString(config.ConfigFlags["keyring-path"]),
		"Path (relative to DefraDB root directory) to store encrypted keys when using the file backend",
	)
	cmd.PersistentFlags().String(
		"secret-file",
		cfg.GetString(config.ConfigFlags["secret-file"]),
		"Path to the file containing secrets")
}
