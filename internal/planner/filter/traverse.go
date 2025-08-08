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
	"slices"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// TraverseFields walks through a filter condition tree and calls the provided function f
// for each leaf node (field value) encountered. The function f receives the path to the field
// (as a string slice) and its value. If f returns false, traversal stops immediately.
//
// The path parameter in f represents the nested field names leading to the value, excluding
// operator keys (those starting with '_'). For example, given the filter:
//
//	{
//	    "author": {
//	        "books": {
//	            "title": {"_eq": "Sample"}
//	        }
//	    }
//	}
//
// The callback would receive path=["author", "books", "title"] and value="Sample"
func TraverseFields(conditions map[string]any, f func([]string, any) bool) {
	traverseFields(nil, "", conditions, f)
}

func traverseFields(path []string, key string, value any, f func([]string, any) bool) bool {
	isKeyOp := func(k string) bool { return len(k) > 0 && k[0] == '_' && k != request.DocIDFieldName }
	isOpComplex := func(k string) bool {
		switch k {
		// all these ops should have a map or an array as value and can not have a single value
		case request.AliasFieldName, request.FilterOpOr, request.FilterOpAnd, request.FilterOpNot:
			return true
		}
		return false
	}
	switch t := value.(type) {
	case map[string]any:
		for key, value := range t {
			if isKeyOp(key) {
				if !traverseFields(path, key, value, f) {
					return false
				}
			} else {
				newPath := make([]string, len(path), len(path)+1)
				copy(newPath, path)
				newPath = append(newPath, key)
				if !traverseFields(newPath, key, value, f) {
					return false
				}
			}
		}
	case []any:
		for _, v := range t {
			if !traverseFields(path, "", v, f) {
				return false
			}
		}
	default:
		if isKeyOp(key) && isOpComplex(key) {
			return false
		}
		return f(path, value)
	}
	return true
}

// TraverseProperties walks through a mapper filter tree and calls the provided function f
// for each PropertyIndex node encountered. Unlike TraverseFields, this function works with
// the internal filter representation using mapper.PropertyIndex and connor.FilterKey types.
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
			shouldIgnore := slices.Contains(skipOps, t.Operation)
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
