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
	"fmt"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/mapper"
	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"
)

type averageNode struct {
	documentIterator
	docMapper

	plan planNode

	sumFieldIndex     int
	countFieldIndex   int
	virtualFieldIndex int
}

func (p *Planner) Average(
	field *mapper.Aggregate,
) (*averageNode, error) {
	var sumField *mapper.Aggregate
	var countField *mapper.Aggregate

	for _, dependency := range field.Dependencies {
		switch dependency.Name {
		case parserTypes.CountFieldName:
			countField = dependency
		case parserTypes.SumFieldName:
			sumField = dependency
		default:
			return nil, errors.New(fmt.Sprintf("Unknown dependency, name: %s", dependency.Name))
		}
	}

	return &averageNode{
		sumFieldIndex:     sumField.Index,
		countFieldIndex:   countField.Index,
		virtualFieldIndex: field.Index,
		docMapper:         docMapper{&field.DocumentMapping},
	}, nil
}

func (n *averageNode) Init() error {
	return n.plan.Init()
}

func (n *averageNode) Kind() string           { return "averageNode" }
func (n *averageNode) Start() error           { return n.plan.Start() }
func (n *averageNode) Spans(spans core.Spans) { n.plan.Spans(spans) }
func (n *averageNode) Close() error           { return n.plan.Close() }
func (n *averageNode) Source() planNode       { return n.plan }

func (n *averageNode) Next() (bool, error) {
	hasNext, err := n.plan.Next()
	if err != nil || !hasNext {
		return hasNext, err
	}

	n.currentValue = n.plan.Value()

	countProp := n.currentValue.Fields[n.countFieldIndex]
	typedCount, isInt := countProp.(int)
	if !isInt {
		return false, errors.New(fmt.Sprintf("Expected count to be int but was: %T", countProp))
	}
	count := typedCount

	if count == 0 {
		n.currentValue.Fields[n.virtualFieldIndex] = float64(0)
		return true, nil
	}

	sumProp := n.currentValue.Fields[n.sumFieldIndex]
	switch sum := sumProp.(type) {
	case float64:
		n.currentValue.Fields[n.virtualFieldIndex] = sum / float64(count)
	case int64:
		n.currentValue.Fields[n.virtualFieldIndex] = float64(sum) / float64(count)
	default:
		return false, errors.New(fmt.Sprintf("Expected sum to be either float64 or int64 or int but was: %T", sumProp))
	}

	return true, nil
}

func (n *averageNode) SetPlan(p planNode) { n.plan = p }

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *averageNode) Explain() (map[string]any, error) {
	return map[string]any{}, nil
}
