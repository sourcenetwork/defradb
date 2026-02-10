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
	"context"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client/options"
)

func MakeP2PActivePeersCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "active-peers",
		Short: "Get list of active peer connections",
		Long: `Get a list of peers that this node is currently connected to.

Results are returned in the multiaddr format (e.g. /ip4/127.0.0.1/tcp/4001/p2p/<PeerID>).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)
			opt := options.WithIdentity(options.ActivePeers(), identity.FromContext(cmd.Context()))
			peers, err := cliClient.ActivePeers(cmd.Context(), opt)
			if err != nil {
				return err
			}
			return writeJSON(cmd, peers)
		},
	}
	return cmd
}
