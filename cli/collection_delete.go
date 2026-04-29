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
	var cmd = &cobra.Command{
		Use:   "delete [collectionNames]",
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
			opt.SetActiveOnly(activeOnly)
			return cliClient.DeleteCollection(cmd.Context(), names, opt)
		},
	}

	cmd.Flags().BoolVar(&activeOnly, "active-only", false,
		"Delete only the active head version of each named collection (default deletes every version)")

	EmbedCLIExample(ctx, cmd, "delete every version of a single collection",
		`defradb client collection delete Users`)

	EmbedCLIExample(ctx, cmd,
		"delete every version of multiple collections in one call (this can be used to delete collections "+
			"that reference each other via relations)",
		`defradb client collection delete Users,Books`)

	EmbedCLIExample(ctx, cmd, "delete only the active head version, keeping earlier versions",
		`defradb client collection delete --active-only Users`)

	return cmd
}
