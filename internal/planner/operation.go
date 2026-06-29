// Copyright 2024 Democratized Data Foundation
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

const operationNodeKind string = "operationNode"

// operationNode is the top level node for operations with
// one or more child selections, such as queries or mutations.
type operationNode struct {
	documentIterator
	docMapper

	// children is indexed by the selection's position in the operation's
	// selection set, so iteration follows the order operations appear in the
	// request. This is required for spec-compliant serial mutation execution
	// (and deterministic default query result order).
	children []planNode
	isDone   bool
}

func (n *operationNode) Prefixes(prefixes []keys.Walkable) {
	for _, child := range n.children {
		child.Prefixes(prefixes)
	}
}

func (n *operationNode) Kind() string {
	return operationNodeKind
}

func (n *operationNode) Init() error {
	n.isDone = false
	n.currentValue = core.Doc{}

	for _, child := range n.children {
		err := child.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *operationNode) Start() error {
	for _, child := range n.children {
		err := child.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *operationNode) Close() error {
	for _, child := range n.children {
		err := child.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *operationNode) Source() planNode {
	return nil
}

func (n *operationNode) Children() []planNode {
	children := make([]planNode, 0, len(n.children))
	for _, child := range n.children {
		children = append(children, child)
	}
	return children
}

func (n *operationNode) Next() (bool, error) {
	if n.isDone {
		return false, nil
	}

	n.currentValue = n.documentMapping.NewDoc()
	for i, child := range n.children {
		switch child.(type) {
		case *topLevelNode:
			hasChild, err := child.Next()
			if err != nil {
				return false, err
			}
			if !hasChild {
				return false, ErrMissingChildValue
			}
			n.currentValue.Fields[i] = child.Value().Fields[0]

		default:
			var docs []core.Doc
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
			n.currentValue.Fields[i] = docs
		}
	}

	n.isDone = true
	return true, nil
}

// Operation creates a new operationNode using the given Selects.
func (p *Planner) Operation(operation *mapper.Operation) (*operationNode, error) {
	// Selection indices are contiguous (0..n-1, assigned by mapper.ToOperation),
	// so a slice sized to the total selection count can be addressed directly by
	// index, preserving request order.
	children := make([]planNode, len(operation.Selects)+len(operation.Mutations)+len(operation.CommitSelects))
	for _, s := range operation.Selects {
		if _, isAgg := request.Aggregates[s.Name]; isAgg {
			// If this Select is an aggregate, then it must be a top-level
			// aggregate and we need to resolve it within the context of a
			// top-level node.
			child, err := p.Top(s)
			if err != nil {
				return nil, err
			}
			children[s.Index] = child
		} else {
			child, err := p.Select(s)
			if err != nil {
				return nil, err
			}
			children[s.Index] = child
		}
	}

	for _, m := range operation.Mutations {
		child, err := p.newObjectMutationPlan(m)
		if err != nil {
			return nil, err
		}
		children[m.Index] = child
	}

	for _, s := range operation.CommitSelects {
		child, err := p.CommitSelect(s)
		if err != nil {
			return nil, err
		}
		children[s.Index] = child
	}

	return &operationNode{
		docMapper: docMapper{operation.DocumentMapping},
		children:  children,
	}, nil
}
