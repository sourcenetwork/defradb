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

	"github.com/ipfs/go-cid"
)

// docSyncResult represents the result of synchronizing a single document.
type docSyncResult struct {
	// DocID is the document ID.
	DocID string
	// Heads is the list of the CID heads for the document.
	Heads []cid.Cid
}

func (p *Peer) SyncDocuments(ctx context.Context, collectionID string, docIDs []string) error {
	_, err := p.server.syncDocuments(ctx, collectionID, docIDs)
	return err
}
