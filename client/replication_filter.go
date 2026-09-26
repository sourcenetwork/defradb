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
	// AllowReplication is called for each incoming document before any of its blocks are
	// stored. A document received without its CAR is checked again once the CAR is fetched. If
	// no peer serves the CAR, only the first check runs.
	//   collectionID: Collection.CollectionID() of the document's collection.
	//   docID:        the document identifier.
	//   fields:       field name to value for the fields written in the document's head commit.
	//                 A value that has not arrived, is encrypted, or does not decode is absent.
	// Return false to discard the document (all its blocks are dropped silently).
	AllowReplication(
		ctx context.Context,
		collectionID string,
		docID string,
		fields map[string]any,
	) bool
}
