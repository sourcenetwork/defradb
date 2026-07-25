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

// Bm25 represents a BM25 relevance score definition.
//
// It holds the target field by name rather than as a Targetable, the way Similarity does,
// because the score is produced by the index scan and nothing reads the field's value.
type Bm25 struct {
	Field

	// Target is the name of the field that is scored. It must carry a BM25 index.
	Target string

	// Query is the text the target field is scored against.
	Query string
}
