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
	"math/big"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

type minNode struct {
	documentIterator
	docMapper

	p      *Planner
	plan   planNode
	parent *mapper.Select

	// virtualFieldIndex is the index of the field
	// that contains the result of the aggregate.
	virtualFieldIndex int
	aggregateMapping  []mapper.AggregateTarget
	aggregateFilter   *mapper.Filter

	execInfo minExecInfo
}

type minExecInfo struct {
	// Total number of times minNode was executed.
	iterations uint64
}

func (p *Planner) Min(
	field *mapper.Aggregate,
	parent *mapper.Select,
	filter *mapper.Filter,
) (*minNode, error) {
	return &minNode{
		p:                 p,
		parent:            parent,
		aggregateMapping:  field.AggregateTargets,
		aggregateFilter:   filter,
		virtualFieldIndex: field.Index,
		docMapper:         docMapper{field.DocumentMapping},
	}, nil
}

func (n *minNode) Kind() string                      { return "minNode" }
func (n *minNode) Init() error                       { return n.plan.Init() }
func (n *minNode) Start() error                      { return n.plan.Start() }
func (n *minNode) Prefixes(prefixes []keys.Walkable) { n.plan.Prefixes(prefixes) }
func (n *minNode) Close() error                      { return n.plan.Close() }
func (n *minNode) Source() planNode                  { return n.plan }
func (n *minNode) SetPlan(p planNode)                { n.plan = p }

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
	for {
		n.execInfo.iterations++

		hasNext, err := n.plan.Next()
		if err != nil || !hasNext {
			return hasNext, err
		}
		n.currentValue = n.plan.Value()

		var min *big.Float
		isFloat := false

		for _, source := range n.aggregateMapping {
			child := n.currentValue.Fields[source.Index]
			var collectionMin *big.Float
			var err error
			switch childCollection := child.(type) {
			case []core.Doc:
				collectionMin = reduceDocs(
					childCollection,
					nil,
					func(childItem core.Doc, value *big.Float) *big.Float {
						childProperty := childItem.Fields[source.ChildTarget.Index]
						res := &big.Float{}
						switch v := childProperty.(type) {
						case int:
							res = res.SetInt64(int64(v))
						case int64:
							res = res.SetInt64(v)
						case uint64:
							res = res.SetUint64(v)
						case float32:
							res = res.SetFloat64(float64(v))
						case float64:
							res = res.SetFloat64(v)
						default:
							return nil
						}
						if value == nil || res.Cmp(value) < 0 {
							return res
						}
						return value
					},
				)

			case []int64:
				collectionMin, err = reduceItems(
					childCollection,
					&source,
					lessN[int64],
					nil,
					func(childItem int64, value *big.Float) *big.Float {
						res := (&big.Float{}).SetInt64(childItem)
						if value == nil || res.Cmp(value) < 0 {
							return res
						}
						return value
					},
				)

			case []immutable.Option[int64]:
				collectionMin, err = reduceItems(
					childCollection,
					&source,
					lessO[int64],
					nil,
					func(childItem immutable.Option[int64], value *big.Float) *big.Float {
						if !childItem.HasValue() {
							return value
						}
						res := (&big.Float{}).SetInt64(childItem.Value())
						if value == nil || res.Cmp(value) < 0 {
							return res
						}
						return value
					},
				)

			case []float64:
				collectionMin, err = reduceItems(
					childCollection,
					&source,
					lessN[float64],
					nil,
					func(childItem float64, value *big.Float) *big.Float {
						res := big.NewFloat(childItem)
						if value == nil || res.Cmp(value) < 0 {
							return res
						}
						return value
					},
				)

			case []immutable.Option[float64]:
				collectionMin, err = reduceItems(
					childCollection,
					&source,
					lessO[float64],
					nil,
					func(childItem immutable.Option[float64], value *big.Float) *big.Float {
						if !childItem.HasValue() {
							return value
						}
						res := big.NewFloat(childItem.Value())
						if value == nil || res.Cmp(value) < 0 {
							return res
						}
						return value
					},
				)
			}
			if err != nil {
				return false, err
			}
			if collectionMin == nil || (min != nil && collectionMin.Cmp(min) >= 0) {
				continue
			}
			isTargetFloat, err := n.p.isValueFloat(n.parent, &source)
			if err != nil {
				return false, err
			}
			isFloat = isTargetFloat
			min = collectionMin
		}

		if min == nil {
			n.currentValue.Fields[n.virtualFieldIndex] = nil
		} else if isFloat {
			res, _ := min.Float64()
			n.currentValue.Fields[n.virtualFieldIndex] = res
		} else {
			res, _ := min.Int64()
			n.currentValue.Fields[n.virtualFieldIndex] = res
		}

		passes, err := mapper.RunFilter(n.currentValue, n.aggregateFilter)
		if err != nil {
			return false, err
		}
		if !passes {
			continue
		}
		return true, nil
	}
}
