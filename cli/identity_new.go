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
	"github.com/sourcenetwork/defradb/crypto"
)

func MakeIdentityNewCommand() *cobra.Command {
	var keyType string

	var cmd = &cobra.Command{
		Use:   "new",
		Short: "Generate a new identity",
		Long: `Generate a new identity

The generated identity contains:
- A private key (secp256k1 or ed25519, based on --type flag)
- A corresponding public key
- A "did:key" generated from the public key.

Example: generate a new identity with secp256k1 key (default):
  defradb identity new

Example: generate a new identity with ed25519 key:
  defradb identity new --type ed25519

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			newIdentity, err := identity.Generate(crypto.KeyType(keyType))
			if err != nil {
				return err
			}

			return writeJSON(cmd, newIdentity)
		},
	}

	cmd.Flags().StringVar(&keyType, "type", "secp256k1", "Key type (secp256k1 or ed25519)")

	return cmd
}
