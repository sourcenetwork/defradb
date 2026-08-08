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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeDocumentPurgeCommand(ctx context.Context) *cobra.Command {
	var docIDs []string
	var pruneHistory bool

	var cmd = &cobra.Command{
		Use:   purgeCommandName,
		Short: "Permanently remove documents by DocID from the local node",
		Long: `Permanently remove a set of documents by DocID from the local node, including all
datastore values and headstore entries. When --prune-history is set, it also removes
reachable blockstore blocks that are no longer owned by another document. Shared blocks
are retained until their final owning document is purged.

History pruning is not supported for branchable collections.

Without --tx, logical cleanup commits one document at a time and can be resumed by
retrying. Each document must fit in one backend transaction. With --tx, the entire purge
must fit in the transaction.

Unlike the soft-delete performed by the delete command, this operation is irreversible and
does not propagate to other nodes in the peer network. It requires the node-level
purge-document permission and does not require collection or document read access.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			parsedDocIDs := make([]client.DocID, 0, len(docIDs))
			for _, raw := range docIDs {
				docID, err := client.NewDocIDFromString(raw)
				if err != nil {
					return NewErrParsingArgument("docID", err)
				}
				parsedDocIDs = append(parsedDocIDs, docID)
			}

			collectionName, _ := cmd.Flags().GetString("collection-name")
			opt := options.WithIdentity(options.PurgeDocuments(), identity.FromContext(cmd.Context()))
			return mustGetContextCLIClient(cmd).PurgeDocuments(
				cmd.Context(),
				collectionName,
				parsedDocIDs,
				pruneHistory,
				opt,
			)
		},
	}

	cmd.Flags().String("collection-name", "", "Collection name")
	_ = cmd.MarkFlagRequired("collection-name")
	cmd.Flags().StringArrayVar(&docIDs, "docID", nil, "DocID of a document to purge (may be repeated)")
	_ = cmd.MarkFlagRequired("docID")
	cmd.Flags().BoolVar(&pruneHistory, "prune-history", false,
		"Also delete reachable blockstore blocks after their final owner is purged")
	EmbedCLIExample(ctx, cmd, "purge a document from the local node",
		`defradb client document purge --collection-name Users --docID bae-123`)
	EmbedCLIExample(ctx, cmd, "purge documents and their unshared history",
		`defradb client document purge --collection-name Users --docID bae-123 --docID bae-456 --prune-history`)

	return cmd
}
