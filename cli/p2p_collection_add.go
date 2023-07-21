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

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/logging"
	netclient "github.com/sourcenetwork/defradb/net/api/client"
)

func MakeP2PCollectionAddCommand(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add [collectionID]",
		Short: "Add P2P collections",
		Long: `Add P2P collections to the synchronized pubsub topics.
The collections are synchronized between nodes of a pubsub network.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
				return errors.New("must specify at least one collectionID")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cred := insecure.NewCredentials()
			client, err := netclient.NewClient(cfg.Net.RPCAddress, grpc.WithTransportCredentials(cred))
			if err != nil {
				return ErrFailedToCreateRPCClient
			}

			rpcTimeoutDuration, err := cfg.Net.RPCTimeoutDuration()
			if err != nil {
				return errors.Wrap("failed to parse RPC timeout duration", err)
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), rpcTimeoutDuration)
			defer cancel()

			err = client.AddP2PCollections(ctx, args...)
			if err != nil {
				return errors.Wrap("failed to add P2P collections, request failed", err)
			}
			log.FeedbackInfo(ctx, "Successfully added P2P collections", logging.NewKV("Collections", args))
			return nil
		},
	}
	return cmd
}
