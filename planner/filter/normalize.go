// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package filter

import (
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

// normalize normalizes the provided filter conditions.
//
// The following cases are subject of normalization:
// - _and or _or with one element is removed flattened
// - double _not is removed
// - any number of consecutive _ands with any number of elements is flattened
//
// As the result object is a map with unique keys (a.k.a. properties),
// while performing flattening of compound operators if the same property
// is present in the result map, both conditions will be moved into an _and
func normalize(conditions map[connor.FilterKey]any) map[connor.FilterKey]any {
	return normalizeCondition(nil, conditions).(map[connor.FilterKey]any)
}

// normalizeCondition returns a normalized version of the given condition.
func normalizeCondition(parentKey connor.FilterKey, condition any) (result any) {
	switch t := condition.(type) {
	case map[connor.FilterKey]any:
		result = normalizeConditions(parentKey, t)

	case []any:
		conditions := make([]any, len(t))
		for i, c := range t {
			conditions[i] = normalizeCondition(parentKey, c)
		}
		result = conditions

	default:
		result = t
	}

	return normalizeProperty(parentKey, result)
}

// normalizeConditions returns a normalized version of the given conditions.
func normalizeConditions(parentKey connor.FilterKey, conditions map[connor.FilterKey]any) map[connor.FilterKey]any {
	result := make(map[connor.FilterKey]any)
	for key, val := range conditions {
		result[key] = normalizeCondition(key, val)

		// check if the condition is an operator that can be normalized
		op, ok := key.(*mapper.Operator)
		if !ok {
			continue
		}
		// check if we have any conditions that can be merged
		merge := normalizeOperator(parentKey, op, result[key])
		if len(merge) == 0 {
			continue
		}
		delete(result, key)

		// merge properties directly into result
		for _, c := range merge {
			for key, val := range c.(map[connor.FilterKey]any) {
				result[key] = val
			}
		}
	}
	return result
}

// normalizeOperator returns a normalized array of conditions.
func normalizeOperator(parentKey connor.FilterKey, op *mapper.Operator, condition any) []any {
	switch op.Operation {
	case request.FilterOpNot:
		return normalizeOperatorNot(condition)

	case request.FilterOpOr:
		return normalizeOperatorOr(condition)

	case request.FilterOpAnd:
		return normalizeOperatorAnd(parentKey, condition)

	default:
		return nil
	}
}

// normalizeOperatorAnd returns an array of conditions with all _and operators merged.
//
// If the parent operator is _not or _or, the subconditions will not be merged.
func normalizeOperatorAnd(parentKey connor.FilterKey, condition any) []any {
	result := condition.([]any)
	// always merge if only 1 property
	if len(result) == 1 {
		return result
	}
	// always merge if parent is not an operator
	parentOp, ok := parentKey.(*mapper.Operator)
	if !ok {
		return result
	}
	// don't merge if parent is a _not or _or operator
	if parentOp.Operation == request.FilterOpNot || parentOp.Operation == request.FilterOpOr {
		return nil
	}
	return result
}

// normalizeOperatorOr returns an array of conditions with all single _or operators merged.
func normalizeOperatorOr(condition any) []any {
	result := condition.([]any)
	// don't merge if more than 1 property
	if len(result) > 1 {
		return nil
	}
	return result
}

// normalizeOperatorNot returns an array of conditions with all double _not operators merged.
func normalizeOperatorNot(condition any) (result []any) {
	subConditions := condition.(map[connor.FilterKey]any)
	// don't merge if more than 1 property
	if len(subConditions) > 1 {
		return nil
	}
	// find double _not occurances
	for subKey, subCondition := range subConditions {
		op, ok := subKey.(*mapper.Operator)
		if ok && op.Operation == request.FilterOpNot {
			result = append(result, subCondition)
		}
	}
	return result
}

// normalizeProperty flattens and groups property filters where possible.
//
// Filters targeting the same property will be grouped into a single _and.
func normalizeProperty(parentKey connor.FilterKey, condition any) any {
	switch t := condition.(type) {
	case map[connor.FilterKey]any:
		results := make(map[connor.FilterKey]any)
		for _, c := range normalizeProperties(parentKey, []any{t}) {
			for key, val := range c.(map[connor.FilterKey]any) {
				results[key] = val
			}
		}
		return results

	case []any:
		return normalizeProperties(parentKey, t)

	default:
		return t
	}
}

// normalizeProperty flattens and groups property filters where possible.
//
// Filters targeting the same property will be grouped into a single _and.
func normalizeProperties(parentKey connor.FilterKey, conditions []any) []any {
	var merge []any
	var result []any

	// can only merge _and groups if parent is not an _or operator
	parentOp, isParentOp := parentKey.(*mapper.Operator)
	canMergeAnd := !isParentOp || parentOp.Operation != request.FilterOpOr

	// accumulate properties that can be merged into a single _and
	// if canMergeAnd is true all _and groups will be merged
	props := make(map[int][]any)
	for _, c := range conditions {
		for key, val := range c.(map[connor.FilterKey]any) {
			op, ok := key.(*mapper.Operator)
			if canMergeAnd && ok && op.Operation == request.FilterOpAnd {
				merge = append(merge, val.([]any)...)
			} else if prop, ok := key.(*mapper.PropertyIndex); ok {
				props[prop.Index] = append(props[prop.Index], map[connor.FilterKey]any{key: val})
			} else {
				result = append(result, map[connor.FilterKey]any{key: val})
			}
		}
	}

	// merge filters with duplicate keys into a single _and
	for _, val := range props {
		if len(val) == 1 {
			// only 1 property so no merge required
			result = append(result, val...)
		} else {
			// multiple properties require merge with _and
			merge = append(merge, val...)
		}
	}

	// nothing to merge
	if len(merge) == 0 {
		return result
	}

	// merge into a single _and operator
	key := &mapper.Operator{Operation: request.FilterOpAnd}
	result = append(result, map[connor.FilterKey]any{key: merge})
	return result
}
