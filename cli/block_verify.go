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

	"github.com/sourcenetwork/defradb/crypto"
)

func MakeBlockVerifySignatureCommand() *cobra.Command {
	var typeStr string
	var cmd = &cobra.Command{
		Args:  cobra.ExactArgs(2),
		Use:   "verify-signature",
		Short: "Verify the signature of a block",
		Long: `Verify the signature of a block by providing the type and public key of the identity.
		
Notes:
  - If 'type' is not provided, secp256k1 is assumed.

Example to verify the signature of a block:
  defradb client block verify-signature --type <type> <public-key> <cid> 
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)
			keyType := crypto.KeyTypeSecp256k1
			if typeStr != "" {
				keyType = crypto.KeyType(typeStr)
			}
			pubKey, err := crypto.PublicKeyFromString(keyType, args[0])
			if err != nil {
				return err
			}
			err = cliClient.VerifySignature(cmd.Context(), args[1], pubKey)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			_, err = out.Write([]byte("Block's signature verified\n"))
			return err
		},
	}
	cmd.Flags().StringVarP(&typeStr, "type", "t", "", "Type of the identity's public key")
	return cmd
}
