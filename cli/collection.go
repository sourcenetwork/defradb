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

	"github.com/sourcenetwork/immutable"
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
)

func MakeCollectionCommand() *cobra.Command {
	var txID uint64
	var name string
	var schemaRoot string
	var versionID string
	var cmd = &cobra.Command{
		Use:   "collection [--name <name> --schema <schemaRoot> --version <versionID>]",
		Short: "Interact with a collection.",
		Long:  `Create, read, update, and delete documents within a collection.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			// cobra does not chain pre run calls so we have to run them again here
			if err := setContextRootDir(cmd); err != nil {
				return err
			}
			if err := setContextConfig(cmd); err != nil {
				return err
			}
			if err := setContextTransaction(cmd, txID); err != nil {
				return err
			}
			if err := setContextStore(cmd); err != nil {
				return err
			}
			store := mustGetContextStore(cmd)

			var col client.Collection
			var cols []client.Collection
			switch {
			case name != "":
				col, err = store.GetCollectionByName(cmd.Context(), name)
				cols = []client.Collection{col}

			default:
				options := client.CollectionFetchOptions{}
				if versionID != "" {
					options.SchemaVersionID = immutable.Some(versionID)
				}
				if schemaRoot != "" {
					options.SchemaRoot = immutable.Some(schemaRoot)
				}

				cols, err = store.GetCollections(cmd.Context(), options)
			}

			if err != nil {
				return err
			}

			if schemaRoot != "" && versionID != "" && len(cols) > 0 {
				if cols[0].SchemaRoot() != schemaRoot {
					// If the a versionID has been provided that does not pair up with the given schema root
					// we should error and let the user know they have provided impossible params.
					// We only need to check the first item - they will all be the same.
					return NewErrSchemaVersionNotOfSchema(schemaRoot, versionID)
				}
			}

			if name != "" {
				// Multiple params may have been specified, and in some cases both are needed.
				// For example if a schema version and a collection name have been provided,
				// we need to ensure that a collection at the requested version is returned.
				// Likewise we need to ensure that if a collection name and schema id are provided,
				// but there are none matching both, that nothing is returned.
				fetchedCols := cols
				cols = nil
				for _, c := range fetchedCols {
					if c.Name().Value() == name {
						cols = append(cols, c)
						break
					}
				}
			}

			if len(cols) != 1 {
				// If more than one collection matches the given criteria we cannot set the context collection
				return nil
			}
			col = cols[0]

			if tx, ok := cmd.Context().Value(txContextKey).(datastore.Txn); ok {
				col = col.WithTxn(tx)
			}

			ctx := context.WithValue(cmd.Context(), colContextKey, col)
			cmd.SetContext(ctx)
			return nil
		},
	}
	cmd.PersistentFlags().Uint64Var(&txID, "tx", 0, "Transaction ID")
	cmd.PersistentFlags().StringVar(&name, "name", "", "Collection name")
	cmd.PersistentFlags().StringVar(&schemaRoot, "schema", "", "Collection schema Root")
	cmd.PersistentFlags().StringVar(&versionID, "version", "", "Collection version ID")
	return cmd
}
