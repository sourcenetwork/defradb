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
	"fmt"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/sourcenetwork/immutable"
	"github.com/spf13/cobra"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
)

func getDefaultNodeIdentityDuration() time.Duration {
	return time.Hour * 24 * 365 * 10 // 10 years
}

func MakeNodeIdentityAssignCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "assign [identity]",
		Short: "Assign an identity to the node",
		Long: `Assign an identity to the node.

Identity is hex-formatted private key.
Node uses the identity to be able to exchange encryption keys with other nodes.
		
Example to assign an identity to the node:
  defradb client node-identity assign 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f 

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("only 1 [identity] argument is allowed")
			}

			cfg := mustGetContextConfig(cmd)

			db := mustGetContextDB(cmd)
			data, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}
			privKey := secp256k1.PrivKeyFromBytes(data)
			identity, err := acpIdentity.FromPrivateKey(
				privKey,
				getDefaultNodeIdentityDuration(),
				immutable.Some(cfg.GetString("api.address")),
				immutable.None[string](),
				false,
			)
			if err != nil {
				return err
			}

			if !cfg.GetBool("keyring.disabled") {
				kr, err := openKeyring(cmd)
				if err != nil {
					return err
				}

				err = kr.Set(nodeIdentityKeyName, data)
				if err != nil {
					return err
				}
			}

			return db.AssignNodeIdentity(cmd.Context(), identity)
		},
	}
	return cmd
}
