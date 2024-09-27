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
	"math"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

type minNode struct {
	documentIterator
	docMapper

	p    *Planner
	plan planNode

	isFloat           bool
	virtualFieldIndex int
	aggregateMapping  []mapper.AggregateTarget

	execInfo minExecInfo
}

type minExecInfo struct {
	// Total number of times minNode was executed.
	iterations uint64
}

func (p *Planner) Min(
	field *mapper.Aggregate,
	parent *mapper.Select,
) (*minNode, error) {
	isFloat := false
	for _, target := range field.AggregateTargets {
		isTargetFloat, err := p.isValueFloat(parent, &target)
		if err != nil {
			return nil, err
		}
		// If one source property is a float, the result will be a float - no need to check the rest
		if isTargetFloat {
			isFloat = true
			break
		}
	}

	return &minNode{
		p:                 p,
		isFloat:           isFloat,
		aggregateMapping:  field.AggregateTargets,
		virtualFieldIndex: field.Index,
		docMapper:         docMapper{field.DocumentMapping},
	}, nil
}

func (n *minNode) Kind() string           { return "minNode" }
func (n *minNode) Init() error            { return n.plan.Init() }
func (n *minNode) Start() error           { return n.plan.Start() }
func (n *minNode) Spans(spans core.Spans) { n.plan.Spans(spans) }
func (n *minNode) Close() error           { return n.plan.Close() }
func (n *minNode) Source() planNode       { return n.plan }
func (n *minNode) SetPlan(p planNode)     { n.plan = p }

func (n *minNode) simpleExplain() (map[string]any, error) {
	sourceExplanations := make([]map[string]any, len(n.aggregateMapping))

	for i, source := range n.aggregateMapping {
		simpleExplainMap := map[string]any{}

		// Add the filter attribute if it exists.
		if source.Filter == nil {
			simpleExplainMap[filterLabel] = nil
		} else {
			// get the target aggregate document mapping. Since the filters
			// are relative to the target aggregate collection (and doc mapper).
			var targetMap *core.DocumentMapping
			if source.Index < len(n.documentMapping.ChildMappings) &&
				n.documentMapping.ChildMappings[source.Index] != nil {
				targetMap = n.documentMapping.ChildMappings[source.Index]
			} else {
				targetMap = n.documentMapping
			}
			simpleExplainMap[filterLabel] = source.Filter.ToMap(targetMap)
		}

		// Add the main field name.
		simpleExplainMap[fieldNameLabel] = source.Field.Name

		// Add the child field name if it exists.
		if source.ChildTarget.HasValue {
			simpleExplainMap[childFieldNameLabel] = source.ChildTarget.Name
		} else {
			simpleExplainMap[childFieldNameLabel] = nil
		}

		sourceExplanations[i] = simpleExplainMap
	}

	return map[string]any{
		sourcesLabel: sourceExplanations,
	}, nil
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *minNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return n.simpleExplain()

	case request.ExecuteExplain:
		return map[string]any{
			"iterations": n.execInfo.iterations,
		}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}

func (n *minNode) Next() (bool, error) {
	n.execInfo.iterations++

	hasNext, err := n.plan.Next()
	if err != nil || !hasNext {
		return hasNext, err
	}
	n.currentValue = n.plan.Value()

	min := math.MaxFloat64
	for _, source := range n.aggregateMapping {
		child := n.currentValue.Fields[source.Index]
		collectionMin := math.MaxFloat64
		var err error
		switch childCollection := child.(type) {
		case []core.Doc:
			collectionMin = reduceDocs(
				childCollection,
				math.MaxFloat64,
				func(childItem core.Doc, value float64) float64 {
					childProperty := childItem.Fields[source.ChildTarget.Index]
					switch v := childProperty.(type) {
					case int:
						return math.Min(value, float64(v))
					case int64:
						return math.Min(value, float64(v))
					case uint64:
						return math.Min(value, float64(v))
					case float64:
						return math.Min(value, float64(v))
					default:
						return value
					}
				},
			)
		case []int64:
			collectionMin, err = reduceItems(
				childCollection,
				&source,
				lessN[int64],
				math.MaxFloat64,
				func(childItem int64, value float64) float64 {
					return math.Min(value, float64(childItem))
				},
			)

		case []immutable.Option[int64]:
			collectionMin, err = reduceItems(
				childCollection,
				&source,
				lessO[int64],
				math.MaxFloat64,
				func(childItem immutable.Option[int64], value float64) float64 {
					if !childItem.HasValue() {
						return value
					}
					return math.Min(value, float64(childItem.Value()))
				},
			)

		case []float64:
			collectionMin, err = reduceItems(
				childCollection,
				&source,
				lessN[float64],
				math.MaxFloat64,
				func(childItem float64, value float64) float64 {
					return math.Min(value, childItem)
				},
			)

		case []immutable.Option[float64]:
			collectionMin, err = reduceItems(
				childCollection,
				&source,
				lessO[float64],
				math.MaxFloat64,
				func(childItem immutable.Option[float64], value float64) float64 {
					if !childItem.HasValue() {
						return value
					}
					return math.Min(value, childItem.Value())
				},
			)
		}
		if err != nil {
			return false, err
		}
		min = math.Min(min, collectionMin)
	}

	if n.isFloat {
		n.currentValue.Fields[n.virtualFieldIndex] = float64(min)
	} else if min >= math.MaxInt64 {
		n.currentValue.Fields[n.virtualFieldIndex] = int64(math.MaxInt64)
	} else {
		n.currentValue.Fields[n.virtualFieldIndex] = int64(min)
	}

	return true, nil
}
