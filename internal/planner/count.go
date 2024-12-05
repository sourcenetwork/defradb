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

// Consider moving this file into an `aggregate` sub-package to keep them organized,
// or moving all aggregates to within an do-all `aggregate` node when adding the next few
// aggregates in.

import (
	"reflect"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

type countNode struct {
	documentIterator
	docMapper

	p    *Planner
	plan planNode

	virtualFieldIndex int
	aggregateMapping  []mapper.AggregateTarget
	aggregateFilter   *mapper.Filter

	execInfo countExecInfo
}

type countExecInfo struct {
	// Total number of times countNode was executed.
	iterations uint64
}

func (p *Planner) Count(field *mapper.Aggregate, host *mapper.Select, filter *mapper.Filter) (*countNode, error) {
	return &countNode{
		p:                 p,
		virtualFieldIndex: field.Index,
		aggregateMapping:  field.AggregateTargets,
		aggregateFilter:   filter,
		docMapper:         docMapper{field.DocumentMapping},
	}, nil
}

func (n *countNode) Kind() string {
	return "countNode"
}

func (n *countNode) Init() error {
	return n.plan.Init()
}

func (n *countNode) Start() error { return n.plan.Start() }

func (n *countNode) Prefixes(prefixes []keys.Walkable) { n.plan.Prefixes(prefixes) }

func (n *countNode) Close() error { return n.plan.Close() }

func (n *countNode) Source() planNode { return n.plan }

func (n *countNode) simpleExplain() (map[string]any, error) {
	sourceExplanations := make([]map[string]any, len(n.aggregateMapping))

	for i, source := range n.aggregateMapping {
		simpleExplainMap := map[string]any{}

		// Add the filter attribute if it exists.
		if source.Filter == nil {
			simpleExplainMap[filterLabel] = nil
		} else {
			// get the target aggregate document mapping. Since the filters
			// are relative to the target aggregate collection (and doc mapper).
			//
			// We can determine if there is a child map if the index from the
			// aggregate target is set (non nil) on the childMapping
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

		sourceExplanations[i] = simpleExplainMap
	}

	return map[string]any{
		sourcesLabel: sourceExplanations,
	}, nil
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *countNode) Explain(explainType request.ExplainType) (map[string]any, error) {
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

func (n *countNode) Next() (bool, error) {
	n.execInfo.iterations++

	hasValue, err := n.plan.Next()
	if err != nil || !hasValue {
		return hasValue, err
	}

	n.currentValue = n.plan.Value()
	// Can just scan for now, can be replaced later by something fancier if needed
	var count int
	for _, source := range n.aggregateMapping {
		property := n.currentValue.Fields[source.Index]
		v := reflect.ValueOf(property)
		switch v.Kind() {
		// v.Len will panic if v is not one of these types, we don't want it to panic
		case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
			if source.Filter == nil && source.Limit == nil {
				count = count + v.Len()
			} else {
				var arrayCount int
				var err error
				switch array := property.(type) {
				case []core.Doc:
					arrayCount = countDocs(array)

				case []bool:
					arrayCount, err = countItems(array, source.Filter, source.Limit)

				case []immutable.Option[bool]:
					arrayCount, err = countItems(array, source.Filter, source.Limit)

				case []int64:
					arrayCount, err = countItems(array, source.Filter, source.Limit)

				case []immutable.Option[int64]:
					arrayCount, err = countItems(array, source.Filter, source.Limit)

				case []float64:
					arrayCount, err = countItems(array, source.Filter, source.Limit)

				case []immutable.Option[float64]:
					arrayCount, err = countItems(array, source.Filter, source.Limit)

				case []string:
					arrayCount, err = countItems(array, source.Filter, source.Limit)

				case []immutable.Option[string]:
					arrayCount, err = countItems(array, source.Filter, source.Limit)
				}
				if err != nil {
					return false, err
				}
				count += arrayCount
			}
		}
	}

	n.currentValue.Fields[n.virtualFieldIndex] = count
	return mapper.RunFilter(n.currentValue, n.aggregateFilter)
}

// countDocs counts the number of documents in a slice, skipping over hidden items
// (a grouping mechanic). Docs should be counted with this function to avoid applying
// offsets twice (once in the select, then once here).
func countDocs(docs []core.Doc) int {
	count := 0
	for _, doc := range docs {
		if !doc.Hidden {
			count += 1
		}
	}

	return count
}

func countItems[T any](source []T, filter *mapper.Filter, limit *mapper.Limit) (int, error) {
	items := enumerable.New(source)
	if filter != nil {
		items = enumerable.Where(items, func(item T) (bool, error) {
			return mapper.RunFilter(item, filter)
		})
	}

	if limit != nil {
		items = enumerable.Skip(items, limit.Offset)
		items = enumerable.Take(items, limit.Limit)
	}

	count := 0
	err := enumerable.OnEach(items, func() {
		count += 1
	})

	return count, err
}

func (n *countNode) SetPlan(p planNode) { n.plan = p }
