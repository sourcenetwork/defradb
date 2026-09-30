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

	"github.com/sourcenetwork/defradb/cli/config"
)

func MakeClientCommand(ctx context.Context) *cobra.Command {
	var txID uint64
	var identity string
	var cmd = &cobra.Command{
		Use:   "client",
		Short: "Interact with a DefraDB node",
		Long: `Interact with a DefraDB node.
Execute queries, add collections, obtain node info, etc.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := setContextRootDir(cmd); err != nil {
				return err
			}
			if err := setContextConfig(cmd); err != nil {
				return err
			}
			if err := setContextIdentity(cmd, identity); err != nil {
				return err
			}
			if err := setContextTransaction(cmd, txID); err != nil {
				return err
			}
			return setContextClient(cmd)
		},
	}
	cmd.PersistentFlags().StringVarP(&identity, "identity", "i", "",
		"Hex formatted private key used to authenticate with ACP")
	cmd.PersistentFlags().Uint64Var(&txID, "tx", 0, "Transaction ID")
	setClientConnectionFlags(cmd)
	return cmd
}

func setClientConnectionFlags(cmd *cobra.Command) {
	cfg := config.DefaultConfig()
	cmd.PersistentFlags().String(
		"url",
		cfg.GetString(config.ConfigFlags["url"]),
		"URL of HTTP endpoint to listen on or connect to",
	)
	cmd.PersistentFlags().String(
		"audience",
		cfg.GetString(config.ConfigFlags["audience"]),
		"Audience to set on minted auth tokens. Defaults to the host of --url",
	)
	cmd.PersistentFlags().String(
		"remote-dac-address",
		cfg.GetString(config.ConfigFlags["remote-dac-address"]),
		"Vera address authorized to make Remote DAC transactions on behalf of the actor",
	)
}
