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
)

func MakeCollectionCommand() *cobra.Command {
	var txID uint64
	var identity string
	var name string
	var collectionID string
	var versionID string
	var getInactive bool
	var cmd = &cobra.Command{
		Use:   "collection [--name <name> --collection-id <collectionID> --version-id <versionID>]",
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

			options := client.CollectionFetchOptions{}
			if versionID != "" {
				options.VersionID = immutable.Some(versionID)
			}
			if collectionID != "" {
				options.CollectionID = immutable.Some(collectionID)
			}
			if name != "" {
				options.Name = immutable.Some(name)
			}
			if getInactive {
				options.IncludeInactive = immutable.Some(getInactive)
			}

			cols, err := cliClient.GetCollections(cmd.Context(), options)
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
	cmd.PersistentFlags().StringVar(&name, "name", "", "Collection name")
	cmd.PersistentFlags().StringVar(&collectionID, "collection-id", "", "Collection ID")
	cmd.PersistentFlags().StringVar(&versionID, "version-id", "", "Collection version ID")
	cmd.PersistentFlags().BoolVar(&getInactive, "get-inactive", false, "Get inactive collections as well as active")
	return cmd
}
