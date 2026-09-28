// Copyright 2025 Democratized Data Foundation
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

// ReplicationFilter allows consumers to filter incoming P2P documents before
// they are stored locally.
type ReplicationFilter interface {
	// AllowReplication is called for each pushed document before any of its blocks are stored.
	// A document received without its CAR is checked first without its field values and again once
	// the CAR is fetched; if no peer serves the CAR, only the first check runs. A refusal at the first
	// check is final, so a filter must allow a document when a field it needs is absent.
	//   collectionID: Collection.CollectionID() of the document's collection. A replicator retry
	//                 passes the collection version ID instead.
	//   docID:        the document identifier.
	//   fields:       field name to CBOR-decoded value for the fields written in the document's
	//                 head commit. A non-negative integer decodes as uint64, and a counter holds its
	//                 increment. A value that has not arrived, is encrypted, or does not decode is
	//                 absent.
	// Return false to discard the document (all its blocks are dropped silently).
	AllowReplication(
		ctx context.Context,
		collectionID string,
		docID string,
		fields map[string]any,
	) bool
}
