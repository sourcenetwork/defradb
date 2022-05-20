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

// DocumentContainer is a specialized buffer to store potentially
// thousands of document value maps. Its used by the Planner system
// to store documents that need to have logic applied to all of them.
// For example, in the sortNode and future groupNode. The Document
// Container acts as an array, so you can append, index, and get the
// length of all the documents inside.
// Close() is called if you want to free all the memory associated
// with the container
type DocumentContainer struct {
	docs    []map[string]interface{}
	numDocs int
}

// NewDocumentContainer returns a new instance of the Document
// Container, with its max buffer size set by capacity.
// A capacity of 0 ignores any initial pre-allocation.
func NewDocumentContainer(capacity int) *DocumentContainer {
	return &DocumentContainer{
		docs:    make([]map[string]interface{}, capacity),
		numDocs: 0,
	}
}

// At returns the document at the specified index.
func (c *DocumentContainer) At(index int) map[string]interface{} {
	if index < 0 || index >= c.numDocs {
		panic("Invalid index for document container")
	}
	return c.docs[index]
}

func (c *DocumentContainer) Len() int {
	return c.numDocs
}

// AddDoc adds a new document to the DocumentContainer.
// It makes a deep copy before its added
func (c *DocumentContainer) AddDoc(doc map[string]interface{}) error {
	if doc == nil {
		return nil
	}
	// append to docs slice
	c.docs = append(c.docs, copyMap(doc))
	c.numDocs++
	return nil
}

// Swap switches the documents at index i and j
// with one another.
func (c *DocumentContainer) Swap(i, j int) {
	if i < 0 || i >= c.numDocs || j < 0 || j >= c.numDocs {
		panic("Invalid index for Document container")
	}

	tmp := c.docs[i]
	c.docs[i] = c.docs[j]
	c.docs[j] = tmp
}

func (c *DocumentContainer) Close() {
	c.docs = nil
	c.numDocs = 0
}

func copyMap(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})
	for k, v := range m {
		vm, ok := v.(map[string]interface{})
		if ok {
			cp[k] = copyMap(vm)
		} else {
			cp[k] = v
		}
	}

	return cp
}
