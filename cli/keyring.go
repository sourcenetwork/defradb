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

func MakeKeyringCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "keyring",
		Short: "Manage DefraDB private keys",
		Long: `Manage DefraDB private keys.
Generate, import, and export private keys.

The following keys are loaded from the keyring on start:
	peer-key: Ed25519 private key (required)
	encryption-key: AES-128, AES-192, or AES-256 key (optional)

To randomly generate the required keys, run the following command:
	defradb keyring generate

To import externally generated keys, run the following command:
	defradb keyring import <name> <private-key-hex>

To learn more about the available options:
	defradb keyring --help
`,
	}
	return cmd
}
