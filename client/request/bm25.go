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

// Bm25 is a functional field holding how well a document matches a text query, scored with BM25
// over one or more of the document's fields.
//
// Unlike Similarity it is not computable from the document alone: the score of a term depends on
// how many documents in the collection contain it and on how long their fields are. Every scored
// field therefore requires a BM25 index on it, and the query errors without one.
type Bm25 struct {
	Field

	// Query is the text the targets are scored against. It is split into terms the same way the
	// indexed values were.
	Query string

	// Targets are the fields scored, in the order they were requested. There is always at least
	// one.
	Targets []Bm25Target
}

// Bm25Target is one field a [Bm25] scores, and how much that field's score counts towards the
// document's total.
type Bm25Target struct {
	// Field is the name of the field scored. It must be a String field carrying a BM25 index.
	Field string

	// Boost multiplies this field's score before it is added to the document's total. It is one
	// unless the request set it, and is never negative.
	//
	// A boost of zero excludes the field: it can neither raise a document's score nor bring a
	// document into the results on its own.
	Boost float64
}
