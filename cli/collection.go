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

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client/options"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

func MakeCollectionCommand(ctx context.Context) *cobra.Command {
	var txID uint64
	var identity string
	var cmd = &cobra.Command{
		Use:   "collection",
		Short: "Interact with a collection.",
		Long:  `Add, describe, patch, set-active, delete, purge-docs, and truncate collections.`,
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

			// The 'add' subcommand creates new collections and doesn't need to resolve an
			// existing collection from the context.
			// If we don't do this, we will hit the NAC gate for collection-get permission
			// when we do the [GetCollections()] call below.
			if cmd.Name() == "add" {
				return nil
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

func setCollectionSelectorFlags(cmd *cobra.Command) {
	cmd.Flags().String("collection-name", "", "Collection name")
	cmd.Flags().String("collection-id", "", "Collection ID")
	cmd.Flags().String("version-id", "", "Collection version ID")
	cmd.Flags().Bool("get-inactive", false, "Get inactive collections as well as active")
}

func getCollectionSelectorOptions(cmd *cobra.Command) *options.GetCollectionsOptionsBuilder {
	name, _ := cmd.Flags().GetString("collection-name")
	collectionID, _ := cmd.Flags().GetString("collection-id")
	versionID, _ := cmd.Flags().GetString("version-id")
	getInactive, _ := cmd.Flags().GetBool("get-inactive")

	opt := options.WithIdentity(options.GetCollections(), iIdentity.FromContext(cmd.Context()))
	if versionID != "" {
		opt.SetVersionID(versionID)
	}
	if collectionID != "" {
		opt.SetCollectionID(collectionID)
	}
	if name != "" {
		opt.SetCollectionName(name)
	}
	if getInactive {
		opt.SetGetInactive(getInactive)
	}
	return opt
}
