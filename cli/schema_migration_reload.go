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

	"github.com/sourcenetwork/defradb/datastore"
)

func MakeSchemaMigrationReloadCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "reload",
		Short: "Reload the schema migrations within DefraDB",
		Long:  `Reload the schema migrations within DefraDB`,
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetContextStore(cmd)

			lens := store.LensRegistry()
			if tx, ok := cmd.Context().Value(txContextKey).(datastore.Txn); ok {
				lens = lens.WithTxn(tx)
			}
			return lens.ReloadLenses(cmd.Context())
		},
	}
	return cmd
}
