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
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

type averageNode struct {
	documentIterator

	plan planNode

	sumFieldName   string
	countFieldName string
	virtualFieldId string
}

func (p *Planner) Average(
	sumField *parser.Select,
	countField *parser.Select,
	field *parser.Select,
) (*averageNode, error) {
	return &averageNode{
		sumFieldName:   sumField.Name,
		countFieldName: countField.Name,
		virtualFieldId: field.Name,
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

	countProp, hasCount := n.currentValue[n.countFieldName]
	sumProp, hasSum := n.currentValue[n.sumFieldName]

	count := 0
	if hasCount {
		typedCount, isInt := countProp.(int)
		if !isInt {
			return false, fmt.Errorf("Expected count to be int but was: %T", countProp)
		}
		count = typedCount
	}

	if count == 0 {
		n.currentValue[n.virtualFieldId] = float64(0)
		return true, nil
	}

	if hasSum {
		switch sum := sumProp.(type) {
		case float64:
			n.currentValue[n.virtualFieldId] = sum / float64(count)
		case int64:
			n.currentValue[n.virtualFieldId] = float64(sum) / float64(count)
		default:
			return false, fmt.Errorf("Expected sum to be either float64 or int64 or int but was: %T", sumProp)
		}
	} else {
		n.currentValue[n.virtualFieldId] = float64(0)
	}

	return true, nil
}

func (n *averageNode) SetPlan(p planNode) { n.plan = p }
