// Copyright 2026 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// sequenceNode is a generic Volcano-model planNode that chains multiple child
// plan nodes sequentially. When the current child is exhausted, it advances
// to the next. This allows composing independent result streams into a single
// ordered pipeline.
//
// Used with orphan handling: ASC ordering puts orphanNode first (NULLs sort first),
// DESC puts it last.
type sequenceNode struct {
	docMapper
	children []planNode
	current  int
}

func newSequenceNode(children ...planNode) *sequenceNode {
	return &sequenceNode{
		children: children,
	}
}

func (n *sequenceNode) Kind() string {
	return "sequenceNode"
}

func (n *sequenceNode) Init() error {
	n.current = 0
	for _, child := range n.children {
		if err := child.Init(); err != nil {
			return err
		}
	}
	return nil
}

func (n *sequenceNode) Start() error {
	for _, child := range n.children {
		if err := child.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (n *sequenceNode) Prefixes(prefixes []keys.Walkable) {
	for _, child := range n.children {
		child.Prefixes(prefixes)
	}
}

func (n *sequenceNode) Next() (bool, error) {
	for n.current < len(n.children) {
		hasNext, err := n.children[n.current].Next()
		if err != nil {
			return false, err
		}
		if hasNext {
			return true, nil
		}
		n.current++
	}
	return false, nil
}

func (n *sequenceNode) Value() core.Doc {
	if n.current < len(n.children) {
		return n.children[n.current].Value()
	}
	return core.Doc{}
}

func (n *sequenceNode) Source() planNode {
	if n.current < len(n.children) {
		return n.children[n.current]
	}
	return nil
}

func (n *sequenceNode) Children() []planNode {
	return n.children
}

func (n *sequenceNode) Close() error {
	var firstErr error
	for _, child := range n.children {
		if err := child.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
