// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package planner

import (
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/container"
)

// A lazily loaded cache-node that allows retrieval of cached documents at arbitrary indexes.
// The node will start empty, and then load items as they are requested.  Items that are
// requested more than once will not be re-loaded from source.
type pipeNode struct {
	source planNode

	docs *container.DocumentContainer

	// The index of the current value - will be -1 if nothinghas been read yet
	docIndex int

	/* Commenting out because unused code (structcheck) according to linter.
	hasNext  bool
	*/
}

func newPipeNode() pipeNode {
	return pipeNode{
		docs: container.NewDocumentContainer(0),
		// A docIndex of minus -1 indicated that nothing has been read yet
		docIndex: -1,
	}
}

func (n *pipeNode) Init() error {
	// We need to make sure state is cleared down on Init,
	// this function may be called multiple times per instance (for example during a join)
	n.docIndex = -1
	n.docs = container.NewDocumentContainer(0)
	return n.source.Init()
}

func (n *pipeNode) Start() error           { return n.source.Start() }
func (n *pipeNode) Spans(spans core.Spans) { n.source.Spans(spans) }
func (n *pipeNode) Close()                 { n.source.Close() }
func (n *pipeNode) Source() planNode       { return n.source }

func (n *pipeNode) Values() map[string]interface{} {
	// Values must be copied out of the node, in case consumers mutate the item
	// for example: when rendering
	return copyMap(n.docs.At(n.docIndex))
}

func (n *pipeNode) Next() (bool, error) {
	// we need to load all docs up until the requested point - this allows us to handle situations where
	// a child record might be requested before handled in the parent - e.g. with a child sort
	for n.docIndex >= n.docs.Len()-1 {
		hasNext, err := n.source.Next()
		if err != nil {
			return false, err
		}
		if !hasNext {
			return false, nil
		}

		doc := n.source.Values()
		n.docs.AddDoc(doc)
	}
	n.docIndex++
	return true, nil
}
