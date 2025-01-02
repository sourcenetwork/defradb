// Copyright 2025 Democratized Data Foundation
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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type traverseNode struct {
	value JSON
	path  string
}

// Helper functions to safely get values
func getObjectValue(j JSON) map[string]JSON {
	if val, ok := j.Value().(map[string]JSON); ok {
		return val
	}
	panic("expected object value")
}

func getArrayValue(j JSON) []JSON {
	if val, ok := j.Value().([]JSON); ok {
		return val
	}
	panic("expected array value")
}

func TestTraverseJSON_ShouldVisitAccordingToConfig(t *testing.T) {
	// Create a complex JSON structure for testing
	json := newJSONObject(map[string]JSON{
		"string": newJSONString("value", nil),
		"number": newJSONNumber(42, nil),
		"bool":   newJSONBool(true, nil),
		"null":   newJSONNull(nil),
		"object": newJSONObject(map[string]JSON{
			"nested": newJSONString("inside", nil),
			"deep": newJSONObject(map[string]JSON{
				"level": newJSONNumber(3, nil),
			}, nil),
		}, nil),
		"array": newJSONArray([]JSON{
			newJSONNumber(1, nil),
			newJSONString("two", nil),
			newJSONObject(map[string]JSON{
				"key": newJSONString("value", nil),
			}, nil),
			newJSONArray([]JSON{
				newJSONNumber(4, nil),
				newJSONNumber(5, nil),
			}, nil),
		}, nil),
	}, nil)

	tests := []struct {
		name     string
		options  []traverseJSONOption
		expected []traverseNode // path -> value
	}{
		{
			name:    "VisitAll",
			options: nil,
			expected: []traverseNode{
				{path: "", value: json},
				{path: "string", value: newJSONString("value", nil)},
				{path: "number", value: newJSONNumber(42, nil)},
				{path: "bool", value: newJSONBool(true, nil)},
				{path: "null", value: newJSONNull(nil)},
				{path: "object", value: getObjectValue(json)["object"]},
				{path: "object/nested", value: newJSONString("inside", nil)},
				{path: "object/deep", value: getObjectValue(getObjectValue(json)["object"])["deep"]},
				{path: "object/deep/level", value: newJSONNumber(3, nil)},
				{path: "array", value: getObjectValue(json)["array"]},
			},
		},
		{
			name: "OnlyLeaves",
			options: []traverseJSONOption{
				TraverseJSONOnlyLeaves(),
			},
			expected: []traverseNode{
				{path: "string", value: newJSONString("value", nil)},
				{path: "number", value: newJSONNumber(42, nil)},
				{path: "bool", value: newJSONBool(true, nil)},
				{path: "null", value: newJSONNull(nil)},
				{path: "object/nested", value: newJSONString("inside", nil)},
				{path: "object/deep/level", value: newJSONNumber(3, nil)},
			},
		},
		{
			name: "WithPrefix_Object",
			options: []traverseJSONOption{
				TraverseJSONWithPrefix([]string{"object"}),
			},
			expected: []traverseNode{
				{path: "object", value: getObjectValue(json)["object"]},
				{path: "object/nested", value: newJSONString("inside", nil)},
				{path: "object/deep", value: getObjectValue(getObjectValue(json)["object"])["deep"]},
				{path: "object/deep/level", value: newJSONNumber(3, nil)},
			},
		},
		{
			name: "WithPrefix_Deep",
			options: []traverseJSONOption{
				TraverseJSONWithPrefix([]string{"object", "deep"}),
			},
			expected: []traverseNode{
				{path: "object/deep", value: getObjectValue(getObjectValue(json)["object"])["deep"]},
				{path: "object/deep/level", value: newJSONNumber(3, nil)},
			},
		},
		{
			name: "VisitArrayElements",
			options: []traverseJSONOption{
				TraverseJSONVisitArrayElements(true),
			},
			expected: []traverseNode{
				{path: "", value: json},
				{path: "string", value: newJSONString("value", nil)},
				{path: "number", value: newJSONNumber(42, nil)},
				{path: "bool", value: newJSONBool(true, nil)},
				{path: "null", value: newJSONNull(nil)},
				{path: "object", value: getObjectValue(json)["object"]},
				{path: "object/nested", value: newJSONString("inside", nil)},
				{path: "object/deep", value: getObjectValue(getObjectValue(json)["object"])["deep"]},
				{path: "object/deep/level", value: newJSONNumber(3, nil)},
				{path: "array", value: getObjectValue(json)["array"]},
				{path: "array", value: newJSONNumber(1, nil)},
				{path: "array", value: newJSONString("two", nil)},
				{path: "array", value: getArrayValue(getObjectValue(json)["array"])[2]},
				{path: "array/key", value: newJSONString("value", nil)},
				{path: "array", value: getArrayValue(getObjectValue(json)["array"])[3]},
				{path: "array", value: newJSONNumber(4, nil)},
				{path: "array", value: newJSONNumber(5, nil)},
			},
		},
		{
			name: "VisitArrayElements without recursion",
			options: []traverseJSONOption{
				TraverseJSONVisitArrayElements(false),
			},
			expected: []traverseNode{
				{path: "", value: json},
				{path: "string", value: newJSONString("value", nil)},
				{path: "number", value: newJSONNumber(42, nil)},
				{path: "bool", value: newJSONBool(true, nil)},
				{path: "null", value: newJSONNull(nil)},
				{path: "object", value: getObjectValue(json)["object"]},
				{path: "object/nested", value: newJSONString("inside", nil)},
				{path: "object/deep", value: getObjectValue(getObjectValue(json)["object"])["deep"]},
				{path: "object/deep/level", value: newJSONNumber(3, nil)},
				{path: "array", value: getObjectValue(json)["array"]},
				{path: "array", value: newJSONNumber(1, nil)},
				{path: "array", value: newJSONString("two", nil)},
			},
		},
		{
			name: "VisitArrayElementsWithIndex",
			options: []traverseJSONOption{
				TraverseJSONVisitArrayElements(true),
				TraverseJSONWithArrayIndexInPath(),
			},
			expected: []traverseNode{
				{path: "", value: json},
				{path: "string", value: newJSONString("value", nil)},
				{path: "number", value: newJSONNumber(42, nil)},
				{path: "bool", value: newJSONBool(true, nil)},
				{path: "null", value: newJSONNull(nil)},
				{path: "object", value: getObjectValue(json)["object"]},
				{path: "object/nested", value: newJSONString("inside", nil)},
				{path: "object/deep", value: getObjectValue(getObjectValue(json)["object"])["deep"]},
				{path: "object/deep/level", value: newJSONNumber(3, nil)},
				{path: "array", value: getObjectValue(json)["array"]},
				{path: "array/0", value: newJSONNumber(1, nil)},
				{path: "array/1", value: newJSONString("two", nil)},
				{path: "array/2", value: getArrayValue(getObjectValue(json)["array"])[2]},
				{path: "array/2/key", value: newJSONString("value", nil)},
				{path: "array/3", value: getArrayValue(getObjectValue(json)["array"])[3]},
				{path: "array/3/0", value: newJSONNumber(4, nil)},
				{path: "array/3/1", value: newJSONNumber(5, nil)},
			},
		},
		{
			name: "CombinedOptions",
			options: []traverseJSONOption{
				TraverseJSONOnlyLeaves(),
				TraverseJSONVisitArrayElements(true),
				TraverseJSONWithPrefix([]string{"array"}),
				TraverseJSONWithArrayIndexInPath(),
			},
			expected: []traverseNode{
				{path: "array/0", value: newJSONNumber(1, nil)},
				{path: "array/1", value: newJSONString("two", nil)},
				{path: "array/2/key", value: newJSONString("value", nil)},
				{path: "array/3/0", value: newJSONNumber(4, nil)},
				{path: "array/3/1", value: newJSONNumber(5, nil)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			visited := []traverseNode{}
			err := TraverseJSON(json, func(value JSON) error {
				key := joinPath(value.GetPath())
				visited = append(visited, traverseNode{path: key, value: value})
				return nil
			}, tt.options...)

			for _, node := range visited {
				if _, ok := node.value.Bool(); ok {
					break
				}
				if _, ok := node.value.Number(); ok {
					break
				}
				if _, ok := node.value.String(); ok {
					break
				}
				if _, ok := node.value.Object(); ok {
					break
				}
				if _, ok := node.value.Array(); ok {
					break
				}
				if node.value.IsNull() {
					break
				}

				t.Errorf("Unexpected JSON value type: %T, for path: %s", node.value, node.path)
			}

			require.NoError(t, err)
			if diff := compareTraverseNodes(tt.expected, visited); diff != "" {
				t.Errorf("Slices are different:\n%s", diff)
			}
		})
	}
}

// compareTraverseNodes compares two slices of traverseNode without relying on the order.
// It matches nodes based on their paths and compares their values.
// Handles multiple nodes with the same path and removes processed items.
func compareTraverseNodes(expected, actual []traverseNode) string {
	var diffs []string

	// Group expected and actual nodes by path
	expectedMap := groupNodesByPath(expected)
	actualMap := groupNodesByPath(actual)

	// Compare nodes with matching paths
	for path, expNodes := range expectedMap {
		actNodes, exists := actualMap[path]
		if !exists {
			diffs = append(diffs, fmt.Sprintf("Missing path %q in actual nodes", path))
			continue
		}

		// Compare each expected node with actual nodes
		for _, expNode := range expNodes {
			matchFound := false
			for i, actNode := range actNodes {
				if compareJSON(expNode.value, actNode.value) {
					// Remove matched node to prevent duplicate matching
					actNodes = append(actNodes[:i], actNodes[i+1:]...)
					actualMap[path] = actNodes
					matchFound = true
					break
				}
			}
			if !matchFound {
				diffs = append(diffs, fmt.Sprintf("No matching value found for path %q", path))
			}
		}

		// Remove path from actualMap if all nodes have been matched
		if len(actNodes) == 0 {
			delete(actualMap, path)
		} else {
			actualMap[path] = actNodes
		}
	}

	// Any remaining actual nodes are extra
	for path, actNodes := range actualMap {
		for range actNodes {
			diffs = append(diffs, fmt.Sprintf("Extra node found at path %q", path))
		}
	}

	if len(diffs) == 0 {
		return ""
	}

	return fmt.Sprintf("Found %d differences:\n%s", len(diffs), strings.Join(diffs, "\n"))
}

// groupNodesByPath groups traverseNodes by their paths.
// It returns a map from path to a slice of nodes with that path.
func groupNodesByPath(nodes []traverseNode) map[string][]traverseNode {
	nodeMap := make(map[string][]traverseNode)
	for _, node := range nodes {
		nodeMap[node.path] = append(nodeMap[node.path], node)
	}
	return nodeMap
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
		"key": newJSONString("value", nil),
	}, nil)

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
