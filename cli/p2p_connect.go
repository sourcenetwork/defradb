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
)

func MakeP2PConnectCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "connect <addresses...>",
		Short: "Connect to one or more peers",
		Long: `Connect to one or more peers with the given addresses

Example: Connect to a peer
  defradb client p2p connect /ip4/0.0.0.0/tcp/9171/p2p/12D3KooW...
  
Example: Connect to multiple peers
  defradb client p2p connect /ip4/0.0.0.0/tcp/9171/p2p/12D3KooW... /ip4/0.0.0.0/tcp/9172/p2p/1543LKs...
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)
			return cliClient.Connect(cmd.Context(), args)
		},
	}
	return cmd
}
