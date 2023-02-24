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
)

var rpcCmd = &cobra.Command{
	Use:   "rpc",
	Short: "Interact with a DefraDB gRPC server",
	Long:  "Interact with a DefraDB gRPC server.",
}

func init() {
	rpcCmd.PersistentFlags().String(
		"addr", cfg.Net.RPCAddress,
		"gRPC endpoint address",
	)
	if err := cfg.BindFlag("net.rpcaddress", rpcCmd.PersistentFlags().Lookup("addr")); err != nil {
		log.FeedbackFatalE(context.Background(), "Could not bind net.rpcaddress", err)
	}
	clientCmd.AddCommand(rpcCmd)
}
