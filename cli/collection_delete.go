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

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeCollectionDeleteCommand(ctx context.Context) *cobra.Command {
	var name string
	var cmd = &cobra.Command{
		Use:   "delete --name <name>",
		Short: "Delete the active collection version",
		Long: `Delete the active version of a collection by name.

Only the latest (head) version is deleted per call. If the collection has multiple
versions, earlier versions must be deleted separately after each head is removed.

The collection must not contain any documents. Delete all documents first before
deleting the collection.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return NewErrRequiredFlagEmpty("name", "n")
			}

			cliClient := mustGetContextCLIClient(cmd)

			opt := options.WithIdentity(options.DeleteCollection(), identity.FromContext(cmd.Context()))
			return cliClient.DeleteCollection(cmd.Context(), name, opt)
		},
	}

	EmbedCLIExample(ctx, cmd, "delete a collection by name",
		`defradb client collection delete --name Users`)

	cmd.Flags().StringVarP(&name, "name", "n", "", "Collection name to delete")

	return cmd
}
