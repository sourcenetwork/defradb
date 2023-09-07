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
	"fmt"
	"reflect"
	"testing"

	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

func assertEqualFilterMap(expected, actual map[connor.FilterKey]any, prefix string) string {
	if len(expected) != len(actual) {
		return fmt.Sprintf("Mismatch at %s: Expected map length: %d, but got: %d", prefix, len(expected), len(actual))
	}

	findMatchingKey := func(key connor.FilterKey, m map[connor.FilterKey]any) connor.FilterKey {
		for k := range m {
			if k.Equal(key) {
				return k
			}
		}
		return nil
	}

	for expKey, expVal := range expected {
		actKey := findMatchingKey(expKey, actual)
		if actKey == nil {
			return fmt.Sprintf("Mismatch at %s: Expected key %v not found in actual map", prefix, expKey)
		}
		actVal := actual[actKey]

		newPrefix := fmt.Sprintf("%s.%v", prefix, expKey)
		switch expTypedVal := expVal.(type) {
		case map[connor.FilterKey]any:
			actTypedVal, ok := actVal.(map[connor.FilterKey]any)
			if !ok {
				return fmt.Sprintf("Mismatch at %s: Expected a nested map[FilterKey]any for key %v, but got: %v", prefix, expKey, actVal)
			}
			errMsg := assertEqualFilterMap(expTypedVal, actTypedVal, newPrefix)
			if errMsg != "" {
				return errMsg
			}
		case []any:
			actTypedVal, ok := actVal.([]any)
			if !ok {
				return fmt.Sprintf("Mismatch at %s: Expected a nested []any for key %v, but got: %v", newPrefix, expKey, actVal)
			}
			if len(expTypedVal) != len(actTypedVal) {
				return fmt.Sprintf("Mismatch at %s: Expected slice length: %d, but got: %d", newPrefix, len(expTypedVal), len(actTypedVal))
			}
			numElements := len(expTypedVal)
			for i := 0; i < numElements; i++ {
				for j := 0; j < numElements; j++ {
					errMsg := compareElements(expTypedVal[i], actTypedVal[j], expKey, newPrefix)
					if errMsg == "" {
						actTypedVal = append(actTypedVal[:j], actTypedVal[j+1:]...)
						break
					}
				}
				if len(actTypedVal) != numElements-i-1 {
					return fmt.Sprintf("Mismatch at %s: Expected element not found: %d", newPrefix, expTypedVal[i])
				}
			}
		default:
			if !reflect.DeepEqual(expVal, actVal) {
				return fmt.Sprintf("Mismatch at %s: Expected value %v for key %v, but got %v", prefix, expVal, expKey, actVal)
			}
		}
	}
	return ""
}

func compareElements(expected, actual any, key connor.FilterKey, prefix string) string {
	switch expElem := expected.(type) {
	case map[connor.FilterKey]any:
		actElem, ok := actual.(map[connor.FilterKey]any)
		if !ok {
			return fmt.Sprintf("Mismatch at %s: Expected a nested map[FilterKey]any for key %v, but got: %v", prefix, key, actual)
		}
		return assertEqualFilterMap(expElem, actElem, prefix)
	default:
		if !reflect.DeepEqual(expElem, actual) {
			return fmt.Sprintf("Mismatch at %s: Expected value %v for key %v, but got %v", prefix, expElem, key, actual)
		}
	}
	return ""
}

func AssertEqualFilterMap(t *testing.T, expected, actual map[connor.FilterKey]any) {
	errMsg := assertEqualFilterMap(expected, actual, "root")
	if errMsg != "" {
		t.Fatal(errMsg)
	}
}

func AssertEqualFilter(t *testing.T, expected, actual *mapper.Filter) {
	if expected == nil && actual == nil {
		return
	}

	if expected == nil || actual == nil {
		t.Fatalf("Expected %v, but got %v", expected, actual)
		return
	}

	AssertEqualFilterMap(t, expected.Conditions, actual.Conditions)

	if !reflect.DeepEqual(expected.ExternalConditions, actual.ExternalConditions) {
		t.Errorf("Expected external conditions \n\t%v\n, but got \n\t%v",
			expected.ExternalConditions, actual.ExternalConditions)
	}
}

func m(op string, val any) map[string]any {
	return map[string]any{op: val}
}

func r(op string, vals ...any) map[string]any {
	return m(op, vals)
}

func getDocMapping() *core.DocumentMapping {
	return &core.DocumentMapping{
		IndexesByName: map[string][]int{"name": {0}, "age": {1}, "published": {2}, "verified": {3}},
		ChildMappings: []*core.DocumentMapping{nil, nil, {
			IndexesByName: map[string][]int{"rating": {11}, "genre": {12}},
		}},
	}
}
