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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeCollectionGetCommand(ctx context.Context) *cobra.Command {
	var showDeleted bool
	var cmd = &cobra.Command{
		Use:   "get [-i --identity] [--show-deleted] <docID> ",
		Short: "View document fields.",
		Long:  `View document fields.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			col, ok := tryGetContextCollection(cmd)
			if !ok {
				return cmd.Usage()
			}

			ctx := cmd.Context()

			docID, err := client.NewDocIDFromString(args[0])
			if err != nil {
				return err
			}

			getOpt := options.WithIdentity(
				options.CollectionGet().SetShowDeleted(showDeleted),
				identity.FromContext(ctx),
			)

			doc, err := col.Get(ctx, docID, getOpt)
			if err != nil {
				return err
			}
			docMap, err := doc.ToMap()
			if err != nil {
				return err
			}
			return writeJSON(cmd, docMap)
		},
	}

	EmbedCLIExample(ctx, cmd, "Get document by ID",
		`defradb client collection get --name User bae-123`)

	EmbedCLIExample(ctx, cmd, "Get a private document using an identity",
		`defradb client collection get --name User bae-123 \
	-i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f `)

	cmd.Flags().BoolVar(&showDeleted, "show-deleted", false, "Show deleted documents")
	return cmd
}
