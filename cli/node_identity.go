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

func MakeNodeIdentityCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "node_identity",
		Short: "Get the information public about the node's identity",
		Long: `Get the information public about the node's identity.

Node uses the identity to be able to exchange encryption keys with other nodes.

The identity contains:
- A compressed 33-byte secp256k1 public key in HEX format.
- A "did:key" generated from the public key.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			db := mustGetContextDB(cmd)
			identity, err := db.GetNodeIdentity(cmd.Context())
			if err != nil {
				return err
			}

			if identity.HasValue() {
				return writeJSON(cmd, identity.Value())
			}

			out := cmd.OutOrStdout()
			_, err = out.Write([]byte("Node has no identity assigned to it\n"))
			return err
		},
	}
	return cmd
}
