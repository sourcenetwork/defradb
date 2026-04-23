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
	var cmd = &cobra.Command{
		Use:   "delete [collectionNames]",
		Short: "Delete the active collection versions",
		Long: `Delete the active version of one or more collections by name.

A single name, or a comma-separated list of names, may be provided. All named
collections are removed atomically in a single operation. This can be used to
delete collections that reference each other via relations, since deleting them
one at a time would leave a dangling reference and be rolled back.

Only the latest (head) version is deleted per call. If a collection has multiple
versions, earlier versions must be deleted separately after each head is removed.

The named collections must not contain any documents. Delete all documents first
before deleting the collection.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var names []string
			for name := range strings.SplitSeq(args[0], ",") {
				name = strings.TrimSpace(name)
				if name == "" {
					continue
				}
				names = append(names, name)
			}

			cliClient := mustGetContextCLIClient(cmd)

			opt := options.WithIdentity(options.DeleteCollection(), identity.FromContext(cmd.Context()))
			return cliClient.DeleteCollection(cmd.Context(), names, opt)
		},
	}

	EmbedCLIExample(ctx, cmd, "delete a single collection",
		`defradb client collection delete Users`)

	EmbedCLIExample(ctx, cmd,
		"delete multiple collections in one call (this can be used to delete collections that reference "+
			"each other via relations)",
		`defradb client collection delete Users,Books`)

	return cmd
}
