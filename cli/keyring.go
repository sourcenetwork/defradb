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

func MakeKeyringCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "keyring",
		Short: "Manage DefraDB private keys",
		Long: `Manage DefraDB private keys.
Generate, add, get, and list private keys.

The following keys are loaded from the keyring on start:
	peer-key: Ed25519 private key (required)
	encryption-key: AES-128, AES-192, or AES-256 key (optional)

To randomly generate the required keys, run the following command:
	defradb keyring new

To import externally generated keys, run the following command:
	defradb keyring add <name> <private-key-hex>

To learn more about the available options:
	defradb keyring --help
`,
	}
	return cmd
}
