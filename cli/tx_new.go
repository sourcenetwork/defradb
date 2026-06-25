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
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
)

type txnTTLClient interface {
	NewTxnWithTTL(bool, time.Duration) (client.Txn, error)
}

func MakeTxNewCommand(ctx context.Context) *cobra.Command {
	var readOnly bool
	var txnTTL string
	var cmd = &cobra.Command{
		Use:   "new",
		Short: "Create a new DefraDB transaction.",
		Long:  `Create a new DefraDB transaction.`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			cliClient := mustGetContextCLIClient(cmd)

			var tx client.Txn
			if txnTTL != "" {
				ttl, err := parseTxnTTLFlag(txnTTL)
				if err != nil {
					return err
				}
				ttlClient, ok := cliClient.(txnTTLClient)
				if !ok {
					return fmt.Errorf("transaction ttl is not supported by this client")
				}
				tx, err = ttlClient.NewTxnWithTTL(readOnly, ttl)
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
	cmd.Flags().StringVar(&txnTTL, "ttl", "", "Transaction idle TTL as a Go duration string, or seconds if no unit is provided")
	return cmd
}

func parseTxnTTLFlag(raw string) (time.Duration, error) {
	txnTTL, err := time.ParseDuration(raw)
	if err == nil {
		return txnTTL, nil
	}
	seconds, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(seconds) * time.Second, nil
}
