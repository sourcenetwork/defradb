// Copyright 2025 Democratized Data Foundation
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

func MakeCollectionPurgeDocsCommand(ctx context.Context) *cobra.Command {
	var docIDStrs []string
	var pruneHistory bool

	var cmd = &cobra.Command{
		Use:   "purge-docs",
		Short: "Permanently remove documents by DocID from the local node",
		Long: `Permanently remove a set of documents by DocID from the local node, including all
datastore values, headstore entries, and, when --prune-history is set, all blockstore
blocks reachable from the documents' head CIDs.

Unlike the soft-delete performed by the delete command, this operation is irreversible and
does not propagate to other nodes in the peer network.`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			col, ok := tryGetContextCollection(cmd)
			if !ok {
				return client.ErrCollectionNotFound
			}

			docIDs := make([]client.DocID, 0, len(docIDStrs))
			for _, raw := range docIDStrs {
				docID, err := client.NewDocIDFromString(raw)
				if err != nil {
					return err
				}
				docIDs = append(docIDs, docID)
			}

			opt := options.WithIdentity(options.TruncateCollection(), identity.FromContext(cmd.Context()))
			return col.PurgeByDocIDs(cmd.Context(), docIDs, pruneHistory, opt)
		},
	}

	setCollectionSelectorFlags(cmd)
	cmd.Flags().StringArrayVar(&docIDStrs, "docID", nil, "DocID of a document to purge (may be repeated)")
	cmd.Flags().BoolVar(&pruneHistory, "prune-history", false,
		"Also delete all blockstore blocks reachable from the documents' head CIDs")

	return cmd
}
