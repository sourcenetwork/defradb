// Copyright 2022 Democratized Data Foundation
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
	"fmt"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/sourcenetwork/defradb/logging"
	netclient "github.com/sourcenetwork/defradb/net/api/client"
)

func MakeAddReplicatorCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "addreplicator",
		Short: "Add a new replicator <collection> <peer>",
		Long: `Use this command if you wish to add a new target replicator
for the p2p data sync system.`,
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			ctx := context.Background()

			collection := args[0]
			peerAddr, err := ma.NewMultiaddr(args[1])
			if err != nil {
				return fmt.Errorf("could not parse peer address: %w", err)
			}

			log.Info(
				ctx,
				"Adding replicator for collection",
				logging.NewKV("PeerAddress", peerAddr),
				logging.NewKV("Collection", collection),
				logging.NewKV("RPCAddress", cfg.Net.RPCAddress))

			cred := insecure.NewCredentials()
			client, err := netclient.NewClient(cfg.Net.RPCAddress, grpc.WithTransportCredentials(cred))
			if err != nil {
				return fmt.Errorf("failed to create RPC client: %w", err)
			}

			rpcTimeoutDuration, err := cfg.Net.RPCTimeoutDuration()
			if err != nil {
				return fmt.Errorf("failed to parse RPC timeout duration: %w", err)
			}
			ctx, cancel := context.WithTimeout(ctx, rpcTimeoutDuration)
			defer cancel()

			pid, err := client.AddReplicator(ctx, collection, peerAddr)
			if err != nil {
				return fmt.Errorf("failed to add replicator, request failed: %w", err)
			}
			log.Info(ctx, "Successfully added replicator", logging.NewKV("PID", pid))
			return nil
		},
	}

	return cmd
}
