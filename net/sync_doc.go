// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

// docSyncRequest represents a request to synchronize specific documents.
type docSyncRequest struct {
	DocIDs []string `json:"docIDs"`
}

// docSyncReply represents the response to a document sync request.
type docSyncReply struct {
	Results []docSyncItem `json:"results"`
	Sender  string        `json:"sender"`
}

// docSyncItem represents the sync result for a single document.
type docSyncItem struct {
	DocID string   `json:"docID"`
	Heads [][]byte `json:"heads"`
}
