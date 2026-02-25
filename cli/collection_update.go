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
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeCollectionUpdateCommand(ctx context.Context) *cobra.Command {
	var argDocID string
	var filter string
	var updater string
	var cmd = &cobra.Command{
		Use:   "update [-i --identity] [--filter <filter> --docID <docID>] --updater <updater>",
		Short: "Update documents by docID or filter.",
		Long:  `Update documents by docID or filter.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			col, ok := tryGetContextCollection(cmd)
			if !ok {
				return cmd.Usage()
			}

			if updater == "" {
				return NewErrMissingRequiredFlag("updater")
			}

			switch {
			case filter != "":
				var filterValue any
				if err := json.Unmarshal([]byte(filter), &filterValue); err != nil {
					return err
				}

				updateWithFilterOpt := options.WithIdentity(
					options.CollectionUpdateWithFilter(), identity.FromContext(ctx))

				res, err := col.UpdateWithFilter(ctx, filterValue, updater, updateWithFilterOpt)
				if err != nil {
					return err
				}
				return writeJSON(cmd, res)
			case argDocID != "":
				docID, err := client.NewDocIDFromString(argDocID)
				if err != nil {
					return err
				}

				getOpt := options.WithIdentity(
					options.CollectionGet().SetShowDeleted(true),
					identity.FromContext(ctx),
				)

				doc, err := col.Get(ctx, docID, getOpt)
				if err != nil {
					return err
				}
				if err := doc.SetWithJSON(ctx, []byte(updater)); err != nil {
					return err
				}

				updateOpt := options.WithIdentity(options.CollectionUpdate(), identity.FromContext(ctx))

				return col.Update(ctx, doc, updateOpt)
			default:
				return ErrNoDocIDOrFilter
			}
		},
	}

	EmbedCLIExample(ctx, cmd, "update by filter",
		`defradb client collection update --name User \
  --filter '{ "points": { "_gte": 100 } }' --updater '{ "verified": true }'`)

	EmbedCLIExample(ctx, cmd, "update by docID",
		`defradb client collection update --name User \
  --docID bae-123 --updater '{ "verified": true }'`)

	EmbedCLIExample(ctx, cmd, "update private docID, with identity",
		`defradb client collection update -i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f --name User \
  --docID bae-123 --updater '{ "verified": true }'`)

	cmd.Flags().StringVar(&argDocID, "docID", "", "Document ID")
	cmd.Flags().StringVar(&filter, "filter", "", "Document filter")
	cmd.Flags().StringVar(&updater, "updater", "", "Document updater")
	return cmd
}
