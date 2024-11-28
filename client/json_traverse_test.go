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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type traverseNode struct {
	value JSON
	path  string
}

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
		expected []traverseNode // path -> value
	}{
		{
			name:    "VisitAll",
			options: nil,
			expected: []traverseNode{
				{path: "", value: json},
				{path: "string", value: newJSONString("value")},
				{path: "number", value: newJSONNumber(42)},
				{path: "bool", value: newJSONBool(true)},
				{path: "null", value: newJSONNull()},
				{path: "object", value: json.Value().(map[string]JSON)["object"]},
				{path: "object/nested", value: newJSONString("inside")},
				{path: "object/deep", value: json.Value().(map[string]JSON)["object"].Value().(map[string]JSON)["deep"]},
				{path: "object/deep/level", value: newJSONNumber(3)},
				{path: "array", value: json.Value().(map[string]JSON)["array"]},
			},
		},
		{
			name: "OnlyLeaves",
			options: []traverseJSONOption{
				TraverseJSONOnlyLeaves(),
			},
			expected: []traverseNode{
				{path: "string", value: newJSONString("value")},
				{path: "number", value: newJSONNumber(42)},
				{path: "bool", value: newJSONBool(true)},
				{path: "null", value: newJSONNull()},
				{path: "object/nested", value: newJSONString("inside")},
				{path: "object/deep/level", value: newJSONNumber(3)},
			},
		},
		{
			name: "WithPrefix_Object",
			options: []traverseJSONOption{
				TraverseJSONWithPrefix([]string{"object"}),
			},
			expected: []traverseNode{
				{path: "object", value: json.Value().(map[string]JSON)["object"]},
				{path: "object/nested", value: newJSONString("inside")},
				{path: "object/deep", value: json.Value().(map[string]JSON)["object"].Value().(map[string]JSON)["deep"]},
				{path: "object/deep/level", value: newJSONNumber(3)},
			},
		},
		{
			name: "WithPrefix_Deep",
			options: []traverseJSONOption{
				TraverseJSONWithPrefix([]string{"object", "deep"}),
			},
			expected: []traverseNode{
				{path: "object/deep", value: json.Value().(map[string]JSON)["object"].Value().(map[string]JSON)["deep"]},
				{path: "object/deep/level", value: newJSONNumber(3)},
			},
		},
		{
			name: "VisitArrayElements",
			options: []traverseJSONOption{
				TraverseJSONVisitArrayElements(),
			},
			expected: []traverseNode{
				{path: "", value: json},
				{path: "string", value: newJSONString("value")},
				{path: "number", value: newJSONNumber(42)},
				{path: "bool", value: newJSONBool(true)},
				{path: "null", value: newJSONNull()},
				{path: "object", value: json.Value().(map[string]JSON)["object"]},
				{path: "object/nested", value: newJSONString("inside")},
				{path: "object/deep", value: json.Value().(map[string]JSON)["object"].Value().(map[string]JSON)["deep"]},
				{path: "object/deep/level", value: newJSONNumber(3)},
				{path: "array", value: json.Value().(map[string]JSON)["array"]},
				{path: "array", value: newJSONNumber(1)},
				{path: "array", value: newJSONString("two")},
				{path: "array", value: json.Value().(map[string]JSON)["array"].Value().([]JSON)[2]},
				{path: "array/key", value: newJSONString("value")},
				{path: "array", value: json.Value().(map[string]JSON)["array"].Value().([]JSON)[3]},
				{path: "array", value: newJSONNumber(4)},
				{path: "array", value: newJSONNumber(5)},
			},
		},
		{
			name: "VisitArrayElementsWithIndex",
			options: []traverseJSONOption{
				TraverseJSONVisitArrayElements(),
				TraverseJSONWithArrayIndexInPath(),
			},
			expected: []traverseNode{
				{path: "", value: json},
				{path: "string", value: newJSONString("value")},
				{path: "number", value: newJSONNumber(42)},
				{path: "bool", value: newJSONBool(true)},
				{path: "null", value: newJSONNull()},
				{path: "object", value: json.Value().(map[string]JSON)["object"]},
				{path: "object/nested", value: newJSONString("inside")},
				{path: "object/deep", value: json.Value().(map[string]JSON)["object"].Value().(map[string]JSON)["deep"]},
				{path: "object/deep/level", value: newJSONNumber(3)},
				{path: "array", value: json.Value().(map[string]JSON)["array"]},
				{path: "array/0", value: newJSONNumber(1)},
				{path: "array/1", value: newJSONString("two")},
				{path: "array/2", value: json.Value().(map[string]JSON)["array"].Value().([]JSON)[2]},
				{path: "array/2/key", value: newJSONString("value")},
				{path: "array/3", value: json.Value().(map[string]JSON)["array"].Value().([]JSON)[3]},
				{path: "array/3/0", value: newJSONNumber(4)},
				{path: "array/3/1", value: newJSONNumber(5)},
			},
		},
		{
			name: "CombinedOptions",
			options: []traverseJSONOption{
				TraverseJSONOnlyLeaves(),
				TraverseJSONVisitArrayElements(),
				TraverseJSONWithPrefix([]string{"array"}),
				TraverseJSONWithArrayIndexInPath(),
			},
			expected: []traverseNode{
				{path: "array/0", value: newJSONNumber(1)},
				{path: "array/1", value: newJSONString("two")},
				{path: "array/2/key", value: newJSONString("value")},
				{path: "array/3/0", value: newJSONNumber(4)},
				{path: "array/3/1", value: newJSONNumber(5)},
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
