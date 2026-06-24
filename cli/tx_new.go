// Copyright 2023 Democratized Data Foundation
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
	"time"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
)

type txnTTLClient interface {
	NewTxnWithTTL(bool, time.Duration) (client.Txn, error)
}

func MakeTxNewCommand(ctx context.Context) *cobra.Command {
	var readOnly bool
	var txnTTL time.Duration
	var cmd = &cobra.Command{
		Use:   "new",
		Short: "Create a new DefraDB transaction.",
		Long:  `Create a new DefraDB transaction.`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cliClient := mustGetContextCLIClient(cmd)

			var tx client.Txn
			if txnTTL != 0 {
				ttlClient, ok := cliClient.(txnTTLClient)
				if !ok {
					return fmt.Errorf("transaction ttl is not supported by this client")
				}
				tx, err = ttlClient.NewTxnWithTTL(readOnly, txnTTL)
			} else {
				tx, err = cliClient.NewTxn(readOnly)
			}
			if err != nil {
				return err
			}
			return writeJSON(cmd, map[string]any{"id": tx.ID()})
		},
	}
	cmd.Flags().BoolVar(&readOnly, "read-only", false, "Transaction is read only")
	cmd.Flags().DurationVar(&txnTTL, "ttl", 0, "Transaction idle TTL")
	return cmd
}
