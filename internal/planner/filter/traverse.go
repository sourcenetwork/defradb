// Copyright 2025 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// OpRelatedFilter stands in for a condition that is not applied to the field itself but to a
// field of the collection it relates to, and so has no operator of its own on this field.
const OpRelatedFilter = ""

// IsOperator reports whether a key in an external filter is an operator rather than a field
// name.
func IsOperator(key string) bool {
	if len(key) == 0 || key[0] != '_' || key == request.DocIDFieldName {
		return false
	}
	_, isRelated := request.ToRelatedObjectName(key)
	return !isRelated
}

// TraverseFieldOperators calls f once per leaf condition in an external filter, with the name
// of the top-level field the condition sits under and the operator applied to that field. A
// condition on a field of a related collection reports OpRelatedFilter, since its operator
// applies to the related field and not to this one.
//
// Conditions under _not are skipped, following what the index fetcher already does with
// them: an index produces candidate documents, and no candidate set can express "every
// document that does not match".
func TraverseFieldOperators(field string, conditions any, f func(field, op string)) {
	switch t := conditions.(type) {
	case map[string]any:
		for key, val := range t {
			switch {
			case !IsOperator(key):
				if field == "" {
					TraverseFieldOperators(key, val, f)
				} else {
					f(field, OpRelatedFilter)
				}
			case key == request.FilterOpNot:
				// Skipped, see the comment above.
			case key == request.FilterOpAnd || key == request.FilterOpOr || key == request.AliasFieldName:
				TraverseFieldOperators(field, val, f)
			case field != "":
				f(field, key)
			}
		}
	case []any:
		for _, val := range t {
			TraverseFieldOperators(field, val, f)
		}
	}
}

// TraverseProperties walks through a mapper filter tree and calls the provided function f
// for each PropertyIndex node encountered. This function works with the internal filter
// representation using mapper.PropertyIndex and connor.FilterKey types.
//
// The function f receives:
// - The property index node being visited
// - A map of its conditions
//
// If f returns false, traversal stops immediately.
func TraverseProperties(
	conditions map[connor.FilterKey]any,
	f func(*mapper.PropertyIndex, map[connor.FilterKey]any) bool,
	skipOps ...string,
) {
	traverseProperties(nil, conditions, f, skipOps)
}

func traverseProperties(
	path []string,
	conditions map[connor.FilterKey]any,
	f func(*mapper.PropertyIndex, map[connor.FilterKey]any) bool,
	skipOps []string,
) bool {
	for filterKey, cond := range conditions {
		switch t := filterKey.(type) {
		case *mapper.PropertyIndex:
			if condMap, ok := cond.(map[connor.FilterKey]any); ok {
				if !f(t, condMap) {
					return false
				}
			}
		case *mapper.Operator:
			// Skip this operator if it's in the ignore list
			shouldIgnore := false
			for _, ignore := range skipOps {
				if t.Operation == ignore {
					shouldIgnore = true
					break
				}
			}
			if shouldIgnore {
				continue
			}

			switch condVal := cond.(type) {
			case map[connor.FilterKey]any:
				if !traverseProperties(path, condVal, f, skipOps) {
					return false
				}
			case []any:
				for _, elem := range condVal {
					if elemMap, ok := elem.(map[connor.FilterKey]any); ok {
						if !traverseProperties(path, elemMap, f, skipOps) {
							return false
						}
					}
				}
			}
		}
	}
	return true
}
