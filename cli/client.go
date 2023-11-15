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
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/config"
)

func MakeClientCommand(cfg *config.Config) *cobra.Command {
	var txID uint64
	var cmd = &cobra.Command{
		Use:   "client",
		Short: "Interact with a DefraDB node",
		Long: `Interact with a DefraDB node.
Execute queries, add schema types, obtain node info, etc.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := loadConfig(cfg); err != nil {
				return err
			}
			if err := setTransactionContext(cmd, cfg, txID); err != nil {
				return err
			}
			return setStoreContext(cmd, cfg)
		},
	}
	cmd.PersistentFlags().Uint64Var(&txID, "tx", 0, "Transaction ID")
	return cmd
}
