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

	"github.com/sourcenetwork/defradb/client"
)

// docSyncResponse represents the response to a document sync request.
type docSyncResponse struct {
	// Results is a slice of sync results with document IDs.
	Results []docSyncResult
	// Sender is the peer ID of the responder.
	Sender string
	// Error is any error that occurred during the sync operation.
	Error error
}

// docSyncResult represents the result of synchronizing a single document.
type docSyncResult struct {
	// DocID is the document ID.
	DocID string
	// Heads is the list of the CID heads for the document.
	Heads []cid.Cid
}

// SyncDocuments requests the latest versions of specified documents from the network
// and synchronizes their DAGs locally. After successful sync, automatically subscribes
// to the documents and their collection for future updates.
func (p *Peer) SyncDocuments(
	ctx context.Context,
	collectionID string,
	docIDs []string,
	opts ...client.DocSyncOption,
) error {
	options := &client.DocSyncOptions{}
	for _, opt := range opts {
		opt(options)
	}

	responseChan := p.server.handleDocSyncRequest(collectionID, docIDs, options.Timeout)

	response := <-responseChan
	return response.Error
}
