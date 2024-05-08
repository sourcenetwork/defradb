// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package container

import (
	"github.com/sourcenetwork/defradb/internal/core"
)

// DocumentContainer is a specialized buffer to store potentially
// thousands of document value maps. Its used by the Planner system
// to store documents that need to have logic applied to all of them.
// For example, in the orderNode and future groupNode. The Document
// Container acts as an array, so you can append, index, and get the
// length of all the documents inside.
// Close() is called if you want to free all the memory associated
// with the container
type DocumentContainer struct {
	docs []core.Doc
}

// NewDocumentContainer returns a new instance of the Document Container, with
// its max buffer size set by capacity. A capacity of 0 ignores any initial pre-allocation.
func NewDocumentContainer(capacity int) *DocumentContainer {
	return &DocumentContainer{
		docs: make([]core.Doc, capacity),
	}
}

// At returns the document at the specified index.
func (c *DocumentContainer) At(index int) core.Doc {
	return c.docs[index]
}

// Len returns the number of documents in the DocumentContainer.
func (c *DocumentContainer) Len() int {
	return len(c.docs)
}

// AddDoc adds a new document to the DocumentContainer.
//
// It makes a deep copy before its added to allow for independent mutation of
// the added clone.
func (c *DocumentContainer) AddDoc(doc core.Doc) {
	copyDoc := doc.Clone()
	c.docs = append(c.docs, copyDoc)
}

// Swap switches the documents at index i and j with one another.
func (c *DocumentContainer) Swap(i, j int) {
	tmp := c.docs[i]
	c.docs[i] = c.docs[j]
	c.docs[j] = tmp
}

// Close frees the DocumentContainer's documents.
func (c *DocumentContainer) Close() {
	c.docs = nil
}
