// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTraverseJSON_ShouldVisitAccordingToConfig(t *testing.T) {
	// Create a complex JSON structure for testing
	json := newJSONObject(map[string]JSON{
		"string": newJSONString("value"),
		"number": newJSONNumber(42),
		"bool":   newJSONBool(true),
		"null":   newJSONNull(),
		"object": newJSONObject(map[string]JSON{
			"nested": newJSONString("inside"),
			"deep": newJSONObject(map[string]JSON{
				"level": newJSONNumber(3),
			}),
		}),
		"array": newJSONArray([]JSON{
			newJSONNumber(1),
			newJSONString("two"),
			newJSONObject(map[string]JSON{
				"key": newJSONString("value"),
			}),
			newJSONArray([]JSON{
				newJSONNumber(4),
				newJSONNumber(5),
			}),
		}),
	})

	tests := []struct {
		name     string
		options  []traverseJSONOption
		expected map[string]JSON // path -> value
	}{
		{
			name:    "VisitAll",
			options: nil,
			expected: map[string]JSON{
				"":                  json,
				"string":            newJSONString("value"),
				"number":            newJSONNumber(42),
				"bool":              newJSONBool(true),
				"null":              newJSONNull(),
				"object":            json.Value().(map[string]JSON)["object"],
				"object/nested":     newJSONString("inside"),
				"object/deep":       json.Value().(map[string]JSON)["object"].Value().(map[string]JSON)["deep"],
				"object/deep/level": newJSONNumber(3),
				"array":             json.Value().(map[string]JSON)["array"],
			},
		},
		{
			name: "OnlyLeaves",
			options: []traverseJSONOption{
				TraverseJSONOnlyLeaves(),
			},
			expected: map[string]JSON{
				"string":            newJSONString("value"),
				"number":            newJSONNumber(42),
				"bool":              newJSONBool(true),
				"null":              newJSONNull(),
				"object/nested":     newJSONString("inside"),
				"object/deep/level": newJSONNumber(3),
			},
		},
		{
			name: "WithPrefix_Object",
			options: []traverseJSONOption{
				TraverseJSONWithPrefix([]string{"object"}),
			},
			expected: map[string]JSON{
				"object":            json.Value().(map[string]JSON)["object"],
				"object/nested":     newJSONString("inside"),
				"object/deep":       json.Value().(map[string]JSON)["object"].Value().(map[string]JSON)["deep"],
				"object/deep/level": newJSONNumber(3),
			},
		},
		{
			name: "WithPrefix_Deep",
			options: []traverseJSONOption{
				TraverseJSONWithPrefix([]string{"object", "deep"}),
			},
			expected: map[string]JSON{
				"object/deep":       json.Value().(map[string]JSON)["object"].Value().(map[string]JSON)["deep"],
				"object/deep/level": newJSONNumber(3),
			},
		},
		{
			name: "VisitArrayElements",
			options: []traverseJSONOption{
				TraverseJSONVisitArrayElements(),
			},
			expected: map[string]JSON{
				"":                  json,
				"string":            newJSONString("value"),
				"number":            newJSONNumber(42),
				"bool":              newJSONBool(true),
				"null":              newJSONNull(),
				"object":            json.Value().(map[string]JSON)["object"],
				"object/nested":     newJSONString("inside"),
				"object/deep":       json.Value().(map[string]JSON)["object"].Value().(map[string]JSON)["deep"],
				"object/deep/level": newJSONNumber(3),
				"array":             json.Value().(map[string]JSON)["array"],
				"array/0":           newJSONNumber(1),
				"array/1":           newJSONString("two"),
				"array/2":           json.Value().(map[string]JSON)["array"].Value().([]JSON)[2],
				"array/2/key":       newJSONString("value"),
				"array/3":           json.Value().(map[string]JSON)["array"].Value().([]JSON)[3],
				"array/3/0":         newJSONNumber(4),
				"array/3/1":         newJSONNumber(5),
			},
		},
		{
			name: "CombinedOptions",
			options: []traverseJSONOption{
				TraverseJSONOnlyLeaves(),
				TraverseJSONVisitArrayElements(),
				TraverseJSONWithPrefix([]string{"array"}),
			},
			expected: map[string]JSON{
				"array/0":     newJSONNumber(1),
				"array/1":     newJSONString("two"),
				"array/2/key": newJSONString("value"),
				"array/3/0":   newJSONNumber(4),
				"array/3/1":   newJSONNumber(5),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visited := make(map[string]JSON)
			err := TraverseJSON(json, func(value JSON) error {
				key := joinPath(value.GetPath())
				visited[key] = value
				return nil
			}, tt.options...)

			require.NoError(t, err)
			if diff := compareJSONMaps(tt.expected, visited); diff != "" {
				t.Errorf("Maps are different:\n%s", diff)
			}
		})
	}
}

// compareJSONMaps compares two maps of JSON values and returns a detailed difference report.
func compareJSONMaps(expected, actual map[string]JSON) string {
	var diffs []string

	// Check for missing keys in actual
	var expectedKeys []string
	for k := range expected {
		expectedKeys = append(expectedKeys, k)
	}
	sort.Strings(expectedKeys)

	var actualKeys []string
	for k := range actual {
		actualKeys = append(actualKeys, k)
	}
	sort.Strings(actualKeys)

	// Find missing keys
	for _, k := range expectedKeys {
		if _, ok := actual[k]; !ok {
			diffs = append(diffs, fmt.Sprintf("- Missing key %q", k))
		}
	}

	// Find extra keys
	for _, k := range actualKeys {
		if _, ok := expected[k]; !ok {
			diffs = append(diffs, fmt.Sprintf("+ Extra key %q", k))
		}
	}

	// Compare values for common keys
	for _, k := range expectedKeys {
		if actualVal, ok := actual[k]; ok {
			expectedVal := expected[k]
			if !compareJSON(expectedVal, actualVal) {
				diffs = append(diffs, fmt.Sprintf("! Value mismatch for key %q:\n\tExpected: %s\n\tActual:   %s",
					k, formatJSON(expectedVal), formatJSON(actualVal)))
			}
		}
	}

	if len(diffs) == 0 {
		return ""
	}

	return fmt.Sprintf("Found %d differences:\n%s", len(diffs), strings.Join(diffs, "\n"))
}

// compareJSON compares two JSON values for equality
func compareJSON(expected, actual JSON) bool {
	if expected.IsNull() != actual.IsNull() {
		return false
	}

	// Compare based on type
	switch {
	case expected.IsNull():
		return true // Both are null (checked above)
	case isObject(expected):
		return compareJSONObjects(expected, actual)
	case isArray(expected):
		return compareJSONArrays(expected, actual)
	default:
		// For primitive types, compare their marshaled form
		expectedBytes, err1 := expected.MarshalJSON()
		actualBytes, err2 := actual.MarshalJSON()
		if err1 != nil || err2 != nil {
			return false
		}
		return string(expectedBytes) == string(actualBytes)
	}
}

func compareJSONObjects(expected, actual JSON) bool {
	expectedObj, ok1 := expected.Object()
	actualObj, ok2 := actual.Object()
	if !ok1 || !ok2 || len(expectedObj) != len(actualObj) {
		return false
	}

	for k, v1 := range expectedObj {
		v2, exists := actualObj[k]
		if !exists || !compareJSON(v1, v2) {
			return false
		}
	}
	return true
}

func compareJSONArrays(expected, actual JSON) bool {
	expectedArr, ok1 := expected.Array()
	actualArr, ok2 := actual.Array()
	if !ok1 || !ok2 || len(expectedArr) != len(actualArr) {
		return false
	}

	for i := range expectedArr {
		if !compareJSON(expectedArr[i], actualArr[i]) {
			return false
		}
	}
	return true
}

// formatJSON returns a human-readable string representation of a JSON value
func formatJSON(j JSON) string {
	switch {
	case j.IsNull():
		return "null"
	case isObject(j):
		obj, _ := j.Object()
		pairs := make([]string, 0, len(obj))
		for k, v := range obj {
			pairs = append(pairs, fmt.Sprintf("%q: %s", k, formatJSON(v)))
		}
		sort.Strings(pairs)
		return "{" + strings.Join(pairs, ", ") + "}"
	case isArray(j):
		arr, _ := j.Array()
		items := make([]string, len(arr))
		for i, v := range arr {
			items[i] = formatJSON(v)
		}
		return "[" + strings.Join(items, ", ") + "]"
	default:
		bytes, _ := j.MarshalJSON()
		return string(bytes)
	}
}

func isObject(j JSON) bool {
	_, ok := j.Object()
	return ok
}

func isArray(j JSON) bool {
	_, ok := j.Array()
	return ok
}

func TestTraverseJSON_WithError(t *testing.T) {
	json := newJSONObject(map[string]JSON{
		"key": newJSONString("value"),
	})

	expectedErr := fmt.Errorf("test error")
	err := TraverseJSON(json, func(value JSON) error {
		return expectedErr
	})

	require.Equal(t, expectedErr, err)
}

func TestShouldVisitPath(t *testing.T) {
	tests := []struct {
		name     string
		prefix   []string
		path     []string
		expected bool
	}{
		{
			name:     "EmptyPrefix",
			prefix:   []string{},
			path:     []string{"a", "b"},
			expected: true,
		},
		{
			name:     "ExactMatch",
			prefix:   []string{"a", "b"},
			path:     []string{"a", "b"},
			expected: true,
		},
		{
			name:     "PrefixMatch",
			prefix:   []string{"a"},
			path:     []string{"a", "b"},
			expected: true,
		},
		{
			name:     "NoMatch",
			prefix:   []string{"a", "b"},
			path:     []string{"a", "c"},
			expected: false,
		},
		{
			name:     "PathTooShort",
			prefix:   []string{"a", "b"},
			path:     []string{"a"},
			expected: true,
		},
		{
			name:     "PathLonger",
			prefix:   []string{"a", "b"},
			path:     []string{"a", "b", "c"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldVisitPath(tt.prefix, tt.path)
			require.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to join path segments
func joinPath(path []string) string {
	if len(path) == 0 {
		return ""
	}
	result := path[0]
	for i := 1; i < len(path); i++ {
		result += "/" + path[i]
	}
	return result
}
