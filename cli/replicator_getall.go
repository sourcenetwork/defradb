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

func MakeReplicatorGetallCommand(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "getall",
		Short: "Get all replicators",
		Long: `Get all the replicators active in the P2P data sync system.
These are the replicators that are currently replicating data from one node to another.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				if err := cmd.Usage(); err != nil {
					return err
				}
				return errors.New("must specify no argument")
			}

			log.FeedbackInfo(
				cmd.Context(),
				"Getting all replicators",
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

			reps, err := client.GetAllReplicators(ctx)
			if err != nil {
				return errors.Wrap("failed to get replicators, request failed", err)
			}
			if len(reps) > 0 {
				log.FeedbackInfo(ctx, "Successfully got all replicators")
				for _, rep := range reps {
					log.FeedbackInfo(
						ctx,
						rep.Info.ID.String(),
						logging.NewKV("Schemas", rep.Schemas),
						logging.NewKV("Addrs", rep.Info.Addrs),
					)
				}
			} else {
				log.FeedbackInfo(ctx, "No replicator found")
			}

			return nil
		},
	}
	return cmd
}
