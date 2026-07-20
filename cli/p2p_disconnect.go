// Copyright 2026 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeP2PDisconnectCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "disconnect <addresses...>",
		Short: "Disconnect from one or more peers",
		Long:  `Disconnect from one or more peers with the given addresses`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)
			opt := options.WithIdentity(options.Disconnect(), identity.FromContext(cmd.Context()))
			return cliClient.Disconnect(cmd.Context(), parseAddresses(args), opt)
		},
	}

	EmbedCLIExample(ctx, cmd, "Disconnect from a peer",
		`defradb client p2p disconnect /ip4/0.0.0.0/tcp/9171/p2p/12D3KooW...`)

	EmbedCLIExample(ctx, cmd, "Disconnect from multiple peers",
		`defradb client p2p disconnect /ip4/0.0.0.0/tcp/9171/p2p/12D3KooW... /ip4/0.0.0.0/tcp/9172/p2p/1543LKs...`)

	return cmd
}
