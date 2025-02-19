// Copyright 2025 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

type similarityNode struct {
	documentIterator
	docMapper

	p    *Planner
	plan planNode

	virtualFieldIndex int
	target            mapper.Targetable
	vector            any
	execInfo          similarityExecInfo
	simFilter         *mapper.Filter
}

type similarityExecInfo struct {
	// Total number of times similarityNode was executed.
	iterations uint64
}

func (p *Planner) Similarity(
	field *mapper.Similarity,
	filter *mapper.Filter,
) *similarityNode {
	return &similarityNode{
		p:                 p,
		virtualFieldIndex: field.Index,
		vector:            field.Vector,
		target:            field.SimilarityTarget,
		simFilter:         filter,
		docMapper:         docMapper{field.DocumentMapping},
	}
}

func (n *similarityNode) Kind() string {
	return "similarityNode"
}

func (n *similarityNode) Init() error {
	return n.plan.Init()
}

func (n *similarityNode) Start() error { return n.plan.Start() }

func (n *similarityNode) Prefixes(prefixes []keys.Walkable) { n.plan.Prefixes(prefixes) }

func (n *similarityNode) Close() error { return n.plan.Close() }

func (n *similarityNode) Source() planNode { return n.plan }

func (n *similarityNode) simpleExplain() (map[string]any, error) {
	simpleExplainMap := map[string]any{}

	simpleExplainMap["vector"] = n.vector
	simpleExplainMap["target"] = n.target.Field.Name

	return map[string]any{
		sourcesLabel: simpleExplainMap,
	}, nil
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *similarityNode) Explain(explainType request.ExplainType) (map[string]any, error) {
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

func (n *similarityNode) Next() (bool, error) {
	for {
		n.execInfo.iterations++

		hasNext, err := n.plan.Next()
		if err != nil || !hasNext {
			return hasNext, err
		}

		n.currentValue = n.plan.Value()

		similarity := float64(0)

		child := n.currentValue.Fields[n.target.Index]
		switch childCollection := child.(type) {
		case []int64:
			vector := convertArray[int64](n.vector)
			result := cosineSimilarity(childCollection, vector)
			similarity = float64(result)
		case []float32:
			vector := convertArray[float32](n.vector)
			result := cosineSimilarity(childCollection, vector)
			similarity = float64(result)
		case []float64:
			vector := convertArray[float64](n.vector)
			result := cosineSimilarity(childCollection, vector)
			similarity = float64(result)
		}

		n.currentValue.Fields[n.virtualFieldIndex] = similarity

		passes, err := mapper.RunFilter(n.currentValue, n.simFilter)
		if err != nil {
			return false, err
		}
		if !passes {
			continue
		}
		return true, nil
	}
}

func (n *similarityNode) SetPlan(p planNode) { n.plan = p }

func cosineSimilarity[T number](
	source []T,
	vector []T,
) T {
	var value T
	for i := range source {
		value += vector[i] * source[i]
	}
	return value
}

func convertArray[T int64 | float32 | float64](val any) []T {
	switch typedVal := val.(type) {
	case []any:
		newArr := make([]T, len(typedVal))
		for i, v := range typedVal {
			newArr[i] = convertToType[T](v)
		}
		return newArr
	}
	return nil
}

func convertToType[T int64 | float32 | float64](val any) T {
	switch v := val.(type) {
	case int64:
		return T(v)
	case float64:
		return T(v)
	case float32:
		return T(v)
	case int8:
		return T(v)
	case int16:
		return T(v)
	case int32:
		return T(v)
	case int:
		return T(v)
	case uint8:
		return T(v)
	case uint16:
		return T(v)
	case uint32:
		return T(v)
	case uint64:
		return T(v)
	case uint:
		return T(v)
	}
	var t T
	return t
}
