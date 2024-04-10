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
)

func MakeClientCommand() *cobra.Command {
	var txID uint64
	var cmd = &cobra.Command{
		Use:   "client",
		Short: "Interact with a DefraDB node",
		Long: `Interact with a DefraDB node.
Execute queries, add schema types, obtain node info, etc.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := setContextRootDir(cmd); err != nil {
				return err
			}
			if err := setContextConfig(cmd); err != nil {
				return err
			}
			return setContextTransaction(cmd, txID)
		},
	}
	cmd.PersistentFlags().Uint64Var(&txID, "tx", 0, "Transaction ID")
	return cmd
}
