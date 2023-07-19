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

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
	netclient "github.com/sourcenetwork/defradb/net/api/client"
)

func MakeReplicatorDeleteCommand(cfg *config.Config) *cobra.Command {
	var (
		fullRep bool
		col     []string
	)
	var cmd = &cobra.Command{
		Use:   "delete [-f, --full | -c, --collection] <peer>",
		Short: "Delete a replicator",
		Long: `Use this command if you wish to remove the target replicator
	for the p2p data sync system.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return errors.New("must specify one argument: PeerID")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			pidString := args[0]

			if len(col) == 0 && !fullRep {
				return errors.New("must run with either --full or --collection")
			}

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
				return errors.Wrap("failed to parse PeerID from string", err)
			}

			err = client.DeleteReplicator(ctx, pid)
			if err != nil {
				return errors.Wrap("failed to delete replicator, request failed", err)
			}
			log.FeedbackInfo(ctx, "Successfully deleted replicator", logging.NewKV("PeerID", pid.String()))
			return nil
		},
	}
	cmd.Flags().BoolVarP(&fullRep, "full", "f", false, "Set the replicator to act on all collections")
	cmd.Flags().StringArrayVarP(&col, "collection", "c",
		[]string{}, "Define the collection for the replicator")
	cmd.MarkFlagsMutuallyExclusive("full", "collection")
	return cmd
}
