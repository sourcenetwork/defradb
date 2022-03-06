// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var (
	rpcAddr    string
	rpcTimeout = 10 * time.Second
)

// clientCmd represents the client command
var rpcCmd = &cobra.Command{
	Use:   "rpc",
	Short: "Interact with a running DefraDB gRPC server",
	Long: `Interact with a running DefraDB gRPC server as a client.
	This command allows you to add replicators and more.`,
}

func init() {
	clientCmd.AddCommand(rpcCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rpcCmd.PersistentFlags().StringVar(&rpcAddr, "addr", "0.0.0.0:9161", "Specify the gRPC endpoint address")
}
