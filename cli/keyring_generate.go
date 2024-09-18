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

	"github.com/sourcenetwork/defradb/crypto"
)

func MakeKeyringGenerateCommand() *cobra.Command {
	var noEncryptionKey bool
	var noPeerKey bool
	var cmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate private keys",
		Long: `Generate private keys.
Randomly generate and store private keys in the keyring.
By default peer and encryption keys will be generated.

The DEFRA_KEYRING_SECRET environment variable must be set to unlock the keyring.
This can also be done with a .env file in the working directory or at a path
defined with the --keyring-secret-file flag.

WARNING: This will overwrite existing keys in the keyring.

Example:
  defradb keyring generate

Example: with no encryption key
  defradb keyring generate --no-encryption-key

Example: with no peer key
  defradb keyring generate --no-peer-key

Example: with system keyring
  defradb keyring generate --keyring-backend system`,
		RunE: func(cmd *cobra.Command, args []string) error {
			keyring, err := openKeyring(cmd)
			if err != nil {
				return err
			}
			if !noEncryptionKey {
				encryptionKey, err := crypto.GenerateAES256()
				if err != nil {
					return err
				}
				err = keyring.Set(encryptionKeyName, encryptionKey)
				if err != nil {
					return err
				}
				log.Info("generated encryption key")
			}
			if !noPeerKey {
				peerKey, err := crypto.GenerateEd25519()
				if err != nil {
					return err
				}
				err = keyring.Set(peerKeyName, peerKey)
				if err != nil {
					return err
				}
				log.Info("generated peer key")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&noEncryptionKey, "no-encryption-key", false,
		"Skip generating an encryption key. Encryption at rest will be disabled")
	cmd.Flags().BoolVar(&noPeerKey, "no-peer-key", false,
		"Skip generating a peer key.")
	return cmd
}
