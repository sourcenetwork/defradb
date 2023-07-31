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

	"github.com/sourcenetwork/defradb/config"
)

func MakeRPCCommand(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "rpc",
		Short: "Interact with a DefraDB node via RPC",
		Long:  "Interact with a DefraDB node via RPC.",
	}
	cmd.PersistentFlags().String(
		"addr", cfg.Net.RPCAddress,
		"RPC endpoint address",
	)

	if err := cfg.BindFlag("net.rpcaddress", cmd.PersistentFlags().Lookup("addr")); err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind net.rpcaddress", err)
	}
	return cmd
}
