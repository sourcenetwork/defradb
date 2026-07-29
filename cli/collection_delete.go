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
	"strings"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeCollectionDeleteCommand(ctx context.Context) *cobra.Command {
	var activeOnly bool
	var collectionNames []string
	var cmd = &cobra.Command{
		Use:   "delete --collection-name <collectionNames>",
		Short: "Delete collections",
		Long: `Delete one or more collections by name.

A single name, or a comma-separated list of names, may be provided. All named
collections are removed atomically in a single operation. This can be used to
delete collections that reference each other via relations, since deleting them
one at a time would leave a dangling reference and be rolled back.

By default, every version of each named collection is deleted (active head and
all earlier versions). Pass --active-only to delete only the latest (head) version
and keep earlier versions intact.

The named collections must not contain any documents. Delete all documents first
before deleting the collection.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			names := make([]string, len(collectionNames))
			for i, name := range collectionNames {
				names[i] = strings.TrimSpace(name)
			}

			cliClient := mustGetContextCLIClient(cmd)

			opt := options.WithIdentity(options.DeleteCollection(), identity.FromContext(cmd.Context()))
			opt.SetActiveOnly(activeOnly)
			return cliClient.DeleteCollection(cmd.Context(), names, opt)
		},
	}

	cmd.Flags().StringSliceVar(&collectionNames, "collection-name", nil, "Collection name")
	cmd.Flags().BoolVar(&activeOnly, "active-only", false,
		"Delete only the active head version of each named collection (default deletes every version)")

	EmbedCLIExample(ctx, cmd, "delete every version of a single collection",
		`defradb client collection delete --collection-name Users`)

	EmbedCLIExample(ctx, cmd,
		"delete every version of multiple collections in one call (this can be used to delete collections "+
			"that reference each other via relations)",
		`defradb client collection delete --collection-name Users,Books`)

	EmbedCLIExample(ctx, cmd, "delete only the active head version, keeping earlier versions",
		`defradb client collection delete --active-only --collection-name Users`)

	return cmd
}
