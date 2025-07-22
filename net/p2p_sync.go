// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"context"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
)

func (p *Peer) SyncDocuments(ctx context.Context, collectionName string, docIDs []string) error {
	clientTxn, err := p.db.NewTxn(ctx, false)
	if err != nil {
		return err
	}
	defer clientTxn.Discard(ctx)

	cols, err := clientTxn.GetCollections(
		ctx,
		client.CollectionFetchOptions{
			Name: immutable.Some(collectionName),
		},
	)
	if err != nil {
		return err
	}
	if len(cols) == 0 {
		return client.NewErrCollectionNotFoundForName(collectionName)
	}

	collectionID := cols[0].Version().CollectionID
	_, err = p.server.syncDocuments(ctx, collectionID, docIDs)
	return err
}
