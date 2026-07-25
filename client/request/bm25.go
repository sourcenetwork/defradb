// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package request

// Bm25 is a functional field holding how well a document's field matches a text query, scored
// with BM25.
//
// Unlike Similarity it is not computable from the document alone: the score of a term depends on
// how many documents in the collection contain it and on how long their fields are. It therefore
// requires a BM25 index on the target field, and the query errors without one.
type Bm25 struct {
	Field

	// Target is the field in the host object that is scored against Query.
	//
	// It must be a String field carrying a BM25 index.
	Target string

	// Query is the text the target field is scored against. It is split into terms the same way
	// the indexed values were.
	Query string
}
