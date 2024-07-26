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
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

const operationNodeKind string = "operationNode"

// operationNode is the top level node for operations with
// one or more child selections, such as queries or mutations.
type operationNode struct {
	documentIterator
	docMapper

	children []planNode
	isDone   bool
}

func (n *operationNode) Spans(spans core.Spans) {
	for _, child := range n.children {
		child.Spans(spans)
	}
}

func (n *operationNode) Kind() string {
	return operationNodeKind
}

func (n *operationNode) Init() error {
	n.isDone = false
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

func (p *operationNode) Children() []planNode {
	return p.children
}

func (n *operationNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return map[string]any{}, nil

	case request.ExecuteExplain:
		return map[string]any{}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
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
			n.currentValue = child.Value()

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
	var children []planNode

	for _, s := range operation.Selects {
		if _, isAgg := request.Aggregates[s.Name]; isAgg {
			// If this Select is an aggregate, then it must be a top-level
			// aggregate and we need to resolve it within the context of a
			// top-level node.
			child, err := p.Top(s)
			if err != nil {
				return nil, err
			}
			children = append(children, child)
		} else {
			child, err := p.Select(s)
			if err != nil {
				return nil, err
			}
			children = append(children, child)
		}
	}

	for _, m := range operation.Mutations {
		child, err := p.newObjectMutationPlan(m)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}

	for _, s := range operation.CommitSelects {
		child, err := p.CommitSelect(s)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}

	if len(children) == 0 {
		return nil, ErrOperationDefinitionMissingSelection
	}

	return &operationNode{
		docMapper: docMapper{operation.DocumentMapping},
		children:  children,
	}, nil
}
