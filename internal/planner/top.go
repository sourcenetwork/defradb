// Copyright 2022 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

const topLevelNodeKind string = "topLevelNode"

// topLevelNode is a special node that represents the very top of the
// plan graph. It has no source, and will only yield a single item
// containing all of its children.
type topLevelNode struct {
	documentIterator
	docMapper

	children     []planNode
	childIndexes []int
	isdone       bool

	// This node's children may use this node as a source
	// this property controls the recursive flow preventing
	// infinate loops.
	isInRecurse bool
}

func (n *topLevelNode) Prefixes(prefixes []keys.Walkable) {
	if n.isInRecurse {
		return
	}
	n.isInRecurse = true
	defer func() {
		n.isInRecurse = false
	}()

	for _, child := range n.children {
		child.Prefixes(prefixes)
	}
}

func (n *topLevelNode) Kind() string {
	return topLevelNodeKind
}

func (n *topLevelNode) Init() error {
	if n.isInRecurse {
		return nil
	}
	n.isInRecurse = true
	defer func() {
		n.isInRecurse = false
	}()

	n.isdone = false
	for _, child := range n.children {
		err := child.Init()
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *topLevelNode) Start() error {
	if n.isInRecurse {
		return nil
	}
	n.isInRecurse = true
	defer func() {
		n.isInRecurse = false
	}()

	for _, child := range n.children {
		err := child.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *topLevelNode) Close() error {
	if n.isInRecurse {
		return nil
	}
	n.isInRecurse = true
	defer func() {
		n.isInRecurse = false
	}()

	for _, child := range n.children {
		err := child.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *topLevelNode) Source() planNode {
	return nil
}

// Children() makes topLevelNode into a MultiNode.
func (p *topLevelNode) Children() []planNode {
	return p.children
}

func (n *topLevelNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return map[string]any{}, nil

	case request.ExecuteExplain:
		return map[string]any{}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}

func (n *topLevelNode) Next() (bool, error) {
	if n.isdone {
		return false, nil
	}

	if n.isInRecurse {
		return true, nil
	}

	n.currentValue = n.documentMapping.NewDoc()
	n.isInRecurse = true
	defer func() {
		n.isInRecurse = false
	}()

	for i, child := range n.children {
		switch child.(type) {
		case *selectTopNode:
			docs := []core.Doc{}
			for {
				hasChild, err := child.Next()
				if err != nil {
					return false, err
				}
				if !hasChild {
					break
				}
				docs = append(docs, child.Value())
			}
			n.currentValue.Fields[n.childIndexes[i]] = docs
		default:
			// This Next will always return a value, as it's source is this node!
			// Even if it adds nothing to the current currentValue, it should still
			// yield it unchanged.
			hasChild, err := child.Next()
			if err != nil {
				return false, err
			}
			if !hasChild {
				return false, ErrMissingChildValue
			}

			n.currentValue = child.Value()
		}
	}

	n.isdone = true
	return true, nil
}

// Top creates a new topLevelNode using the given Select.
func (p *Planner) Top(m *mapper.Select) (*topLevelNode, error) {
	node := topLevelNode{
		docMapper: docMapper{m.DocumentMapping},
	}

	aggregateChildren := []planNode{}
	aggregateChildIndexes := []int{}
	for _, field := range m.Fields {
		switch f := field.(type) {
		case *mapper.Aggregate:
			var child planNode
			var err error
			switch field.GetName() {
			case request.CountFieldName:
				child, err = p.Count(f, m, nil)
			case request.SumFieldName:
				child, err = p.Sum(f, m, nil)
			case request.AverageFieldName:
				child, err = p.Average(f, nil)
			case request.MaxFieldName:
				child, err = p.Max(f, m, nil)
			case request.MinFieldName:
				child, err = p.Min(f, m, nil)
			}
			if err != nil {
				return nil, err
			}
			aggregateChildren = append(aggregateChildren, child)
			aggregateChildIndexes = append(aggregateChildIndexes, field.GetIndex())

		case *mapper.Select:
			child, err := p.Select(f)
			if err != nil {
				return nil, err
			}
			node.children = append(node.children, child)
			node.childIndexes = append(node.childIndexes, field.GetIndex())
		}
	}

	// Iterate through the aggregates backwards to ensure dependencies
	// execute *before* any aggregate dependent on them.
	for i := len(aggregateChildren) - 1; i >= 0; i-- {
		node.children = append(node.children, aggregateChildren[i])
		node.childIndexes = append(node.childIndexes, aggregateChildIndexes[i])
	}

	return &node, nil
}
