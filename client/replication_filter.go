// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import "context"

// ReplicationFilter is an interface that allows consumers to filter incoming
// P2P documents before they are stored. When a filter is registered, each
// incoming document's field values are extracted and passed to AllowReplication.
// Returning false causes the document to be silently skipped.
type ReplicationFilter interface {
	// AllowReplication decides whether an incoming P2P document should be stored.
	// collectionID is the DefraDB collection identifier.
	// docID is the document's unique identifier.
	// fields contains the document's field name-value pairs extracted from the
	// incoming CRDT deltas. Values are decoded from CBOR.
	// Returning false causes the document (and its blocks) to be skipped.
	AllowReplication(ctx context.Context, collectionID string, docID string, fields map[string]any) bool
}
