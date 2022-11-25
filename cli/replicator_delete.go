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

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
	netclient "github.com/sourcenetwork/defradb/net/api/client"
)

var deleteReplicatorCmd = &cobra.Command{
	Use:   "delete <peer>",
	Short: "Delete a replicator",
	Long: `Use this command if you wish to remove the target replicator
for the p2p data sync system.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			if err := cmd.Usage(); err != nil {
				return err
			}
			return errors.New("must specify one argument: peer")
		}
		pidString := args[0]

		log.FeedbackInfo(
			cmd.Context(),
			"Removing replicator",
			logging.NewKV("PeerID", pidString),
			logging.NewKV("RPCAddress", cfg.Net.RPCAddress),
		)

		cred := insecure.NewCredentials()
		client, err := netclient.NewClient(cfg.Net.RPCAddress, grpc.WithTransportCredentials(cred))
		if err != nil {
			return errors.Wrap("failed to create RPC client", err)
		}

		rpcTimeoutDuration, err := cfg.Net.RPCTimeoutDuration()
		if err != nil {
			return errors.Wrap("failed to parse RPC timeout duration", err)
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), rpcTimeoutDuration)
		defer cancel()

		pid, err := peer.Decode(pidString)
		if err != nil {
			return errors.Wrap("failed to parse peer id from string", err)
		}

		err = client.DeleteReplicator(ctx, pid)
		if err != nil {
			return errors.Wrap("failed to delete replicator, request failed", err)
		}
		log.FeedbackInfo(ctx, "Successfully deleted replicator", logging.NewKV("PID", pid.String()))
		return nil
	},
}

func init() {
	replicatorCmd.AddCommand(deleteReplicatorCmd)
}
