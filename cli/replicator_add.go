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

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
	netclient "github.com/sourcenetwork/defradb/net/api/client"
)

var (
	fullRep bool
	col     []string
)

var addReplicatorCmd = &cobra.Command{
	Use:   "add [-f, --full | -c, --collection] <peer>",
	Short: "Add a new replicator",
	Long: `Use this command if you wish to add a new target replicator
for the p2p data sync system.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return errors.New("must specify one argument: peer")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(col, args)
		peerAddr, err := ma.NewMultiaddr(args[0])
		if err != nil {
			return errors.Wrap("could not parse peer address", err)
		}

		if len(col) != 0 {
			log.FeedbackInfo(
				cmd.Context(),
				"Adding replicator for collection",
				logging.NewKV("PeerAddress", peerAddr),
				logging.NewKV("Collection", col),
				logging.NewKV("RPCAddress", cfg.Net.RPCAddress),
			)
		} else {
			if !fullRep {
				return errors.New("must run with either --full or --collection")
			}
			log.FeedbackInfo(
				cmd.Context(),
				"Adding full replicator",
				logging.NewKV("PeerAddress", peerAddr),
				logging.NewKV("RPCAddress", cfg.Net.RPCAddress),
			)
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

		pid, err := client.AddReplicator(ctx, peerAddr, col...)
		if err != nil {
			return errors.Wrap("failed to add replicator, request failed", err)
		}
		log.FeedbackInfo(ctx, "Successfully added replicator", logging.NewKV("PID", pid))
		return nil
	},
}

func init() {
	replicatorCmd.AddCommand(addReplicatorCmd)
	addReplicatorCmd.Flags().BoolVarP(&fullRep, "full", "f", false, "Set the replicator to act on all collections")
	addReplicatorCmd.Flags().StringArrayVarP(&col, "collection", "c", []string{}, "Define the collection for the replicator")
	addReplicatorCmd.MarkFlagsMutuallyExclusive("full", "collection")
}
