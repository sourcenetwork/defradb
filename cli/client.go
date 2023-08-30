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

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/http"
	"github.com/spf13/cobra"
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
			db, err := http.NewClient(cfg.API.Address)
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			if txID != 0 {
				ctx = context.WithValue(ctx, storeContextKey, db.WithTxnID(txID))
			} else {
				ctx = context.WithValue(ctx, storeContextKey, db)
			}
			ctx = context.WithValue(ctx, dbContextKey, db)
			cmd.SetContext(ctx)
			return nil
		},
	}
	cmd.PersistentFlags().Uint64Var(&txID, "tx", 0, "Transaction ID")
	return cmd
}
