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
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
)

// SyncDocuments requests the latest versions of specified documents from the network
// and synchronizes their DAGs locally. After successful sync, automatically subscribes
// to the documents and their collection for future updates.
func (p *Peer) SyncDocuments(
	ctx context.Context,
	collectionID string,
	docIDs []string,
	opts ...client.DocSyncOption,
) <-chan error {
	options := &client.DocSyncOptions{
		Timeout: 10 * time.Second,
	}
	for _, opt := range opts {
		opt(options)
	}

	responseChan := make(chan event.DocSyncResponse, 1)

	request := event.DocSyncRequest{
		CollectionID: collectionID,
		DocIDs:       docIDs,
		Timeout:      options.Timeout,
		Response:     responseChan,
	}

	p.bus.Publish(event.NewMessage(event.DocSyncRequestName, request))

	resultChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		response := <-responseChan
		if response.Error != nil {
			resultChan <- response.Error
		} else {
			resultChan <- nil
		}
	}()

	return resultChan
}
