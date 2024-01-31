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
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

func MakeTxCreateCommand() *cobra.Command {
	var concurrent bool
	var readOnly bool
	var cmd = &cobra.Command{
		Use:   "create",
		Short: "Create a new DefraDB transaction.",
		Long:  `Create a new DefraDB transaction.`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			db := cmd.Context().Value(dbContextKey).(client.DB)

			var tx datastore.Txn
			if concurrent {
				tx, err = db.NewConcurrentTxn(cmd.Context(), readOnly)
			} else {
				tx, err = db.NewTxn(cmd.Context(), readOnly)
			}
			if err != nil {
				return err
			}
			return writeJSON(cmd, map[string]any{"id": tx.ID()})
		},
	}
	cmd.Flags().BoolVar(&concurrent, "concurrent", false, "Transaction is concurrent")
	cmd.Flags().BoolVar(&readOnly, "read-only", false, "Transaction is read only")
	return cmd
}
