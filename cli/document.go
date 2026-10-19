// Copyright 2026 Democratized Data Foundation
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
)

func MakeDocumentCommand(ctx context.Context) *cobra.Command {
	var txID uint64
	var identity string
	var cmd = &cobra.Command{
		Use:   "document",
		Short: "Interact with documents.",
		Long:  `Add, read, update, and delete documents within a collection.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			// cobra does not chain pre run calls so we have to run them again here
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
			if err := setContextClient(cmd); err != nil {
				return err
			}
			cliClient := mustGetContextCLIClient(cmd)
			opt := getCollectionSelectorOptions(cmd)
			cols, err := cliClient.GetCollections(cmd.Context(), opt)
			if err != nil {
				return err
			}

			if len(cols) != 1 {
				// If more than one collection matches the given criteria we cannot set the context collection
				return nil
			}
			col := cols[0]

			ctx := context.WithValue(cmd.Context(), colContextKey, col)
			cmd.SetContext(ctx)
			return nil
		},
	}
	cmd.PersistentFlags().Uint64Var(&txID, "tx", 0, "Transaction ID")
	cmd.PersistentFlags().StringVarP(&identity, "identity", "i", "",
		"Hex formatted private key used to authenticate with ACP")
	return cmd
}
