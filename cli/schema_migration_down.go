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
	"encoding/json"

	"github.com/sourcenetwork/immutable/enumerable"
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/datastore"
)

func MakeSchemaMigrationDownCommand() *cobra.Command {
	var schemaVersionID string
	var cmd = &cobra.Command{
		Use:   "down --version <version> <documents>",
		Short: "Reverses the migration from the specified schema version.",
		Long: `Reverses the migration from the specified schema version.
Documents is a list of documents to reverse the migration from.

Example:
  defradb client schema migration down --version bae123 '[{"name": "Bob"}]'
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := mustGetStoreContext(cmd)

			var src []map[string]any
			if err := json.Unmarshal([]byte(args[0]), &src); err != nil {
				return err
			}
			lens := store.LensRegistry()
			if tx, ok := cmd.Context().Value(txContextKey).(datastore.Txn); ok {
				lens = lens.WithTxn(tx)
			}
			out, err := lens.MigrateDown(cmd.Context(), enumerable.New(src), schemaVersionID)
			if err != nil {
				return err
			}
			return writeJSON(cmd, out)
		},
	}
	cmd.Flags().StringVar(&schemaVersionID, "version", "", "Schema version id")
	return cmd
}
