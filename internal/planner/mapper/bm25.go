// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mapper

import "github.com/sourcenetwork/defradb/client/request"

// Bm25 represents a BM25 relevance score definition.
//
// It holds the scored fields by name rather than as Targetables, the way Similarity does, because
// the score is produced by the index scan and nothing reads the fields' values. For the same
// reason the targets are [request.Bm25Target] unchanged: there is no document mapping to apply to
// them.
type Bm25 struct {
	Field

	// Query is the text the targets are scored against.
	Query string

	// Targets are the fields scored, each with the weight its score is given. Every one of them
	// must carry a BM25 index.
	Targets []request.Bm25Target
}
