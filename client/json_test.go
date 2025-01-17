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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fastjson"
)

func TestParseJSONAndMarshal_WithValidInput_ShouldMarshal(t *testing.T) {
	tests := []struct {
		name     string
		fromFunc func(string) (JSON, error)
	}{
		{
			name:     "FromBytes",
			fromFunc: func(data string) (JSON, error) { return ParseJSONBytes([]byte(data)) },
		},
		{
			name:     "FromString",
			fromFunc: ParseJSONString,
		},
		{
			name: "FromFastJSON",
			fromFunc: func(data string) (JSON, error) {
				var p fastjson.Parser
				v, err := p.Parse(data)
				if err != nil {
					return nil, err
				}
				return NewJSONFromFastJSON(v), nil
			},
		},
		{
			name: "FromMap",
			fromFunc: func(data string) (JSON, error) {
				var result map[string]any
				if err := json.Unmarshal([]byte(data), &result); err != nil {
					return nil, err
				}
				return NewJSONFromMap(result)
			},
		},
	}

	data := `{"key1": "value1", "key2": 2, "key3": true, "key4": null, "key5": ["item1", 2, false]}`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonObj, err := tt.fromFunc(data)
			require.NoError(t, err, "fromFunc failed with error %v", err)

			var buf bytes.Buffer
			err = jsonObj.Marshal(&buf)
			require.NoError(t, err, "jsonObj.Marshal(&buf) failed with error %v", err)

			actualStr := strings.ReplaceAll(buf.String(), "\n", "")
			expectedStr := strings.ReplaceAll(data, " ", "")
			require.Equal(t, actualStr, expectedStr, "Expected %s, got %s", expectedStr, actualStr)

			rawJSON, err := jsonObj.MarshalJSON()
			require.NoError(t, err, "jsonObj.MarshalJSON() failed with error %v", err)
			actualStr = strings.ReplaceAll(string(rawJSON), "\n", "")
			require.Equal(t, actualStr, expectedStr, "Expected %s, got %s", expectedStr, actualStr)
		})
	}
}

func TestNewJSONAndMarshal_WithInvalidInput_ShouldFail(t *testing.T) {
	tests := []struct {
		name     string
		fromFunc func(string) (JSON, error)
	}{
		{
			name:     "FromBytes",
			fromFunc: func(data string) (JSON, error) { return ParseJSONBytes([]byte(data)) },
		},
		{
			name:     "FromString",
			fromFunc: ParseJSONString,
		},
		{
			name: "FromMap",
			fromFunc: func(data string) (JSON, error) {
				var result map[string]any
				if err := json.Unmarshal([]byte(data), &result); err != nil {
					return nil, err
				}
				return NewJSONFromMap(result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.fromFunc(`{"key1": "value1}`)
			require.Error(t, err, "Expected error, but got nil")
		})
	}
}

func TestNewJSONFomString_WithInvalidInput_Error(t *testing.T) {
	_, err := ParseJSONString("str")
	require.Error(t, err, "Expected error, but got nil")
}

func TestJSONObject_Methods_ShouldWorkAsExpected(t *testing.T) {
	m := map[string]JSON{
		"key": newJSONString("value", nil),
		"nested": newJSONObject(map[string]JSON{
			"inner": newJSONNumber(42, nil),
			"array": newJSONArray([]JSON{newJSONString("test", nil), newJSONBool(true, nil)}, nil),
		}, nil),
	}
	obj := newJSONObject(m, nil)
	expectedUnwrapped := map[string]any{
		"key": "value",
		"nested": map[string]any{
			"inner": float64(42),
			"array": []any{"test", true},
		},
	}

	// Positive tests
	val, ok := obj.Object()
	require.True(t, ok)
	require.Equal(t, m, val)
	require.Equal(t, m, obj.Value())
	require.Equal(t, expectedUnwrapped, obj.Unwrap())

	// Negative tests
	_, ok = obj.Array()
	require.False(t, ok)
	_, ok = obj.Number()
	require.False(t, ok)
	_, ok = obj.String()
	require.False(t, ok)
	_, ok = obj.Bool()
	require.False(t, ok)
	require.False(t, obj.IsNull())
}

func TestJSONArray_Methods_ShouldWorkAsExpected(t *testing.T) {
	arr := []JSON{
		newJSONString("item1", nil),
		newJSONObject(map[string]JSON{
			"key": newJSONString("value", nil),
			"num": newJSONNumber(42, nil),
		}, nil),
		newJSONNumber(2, nil),
	}
	jsonArr := newJSONArray(arr, nil)
	expectedUnwrapped := []any{
		"item1",
		map[string]any{
			"key": "value",
			"num": float64(42),
		},
		float64(2),
	}

	// Positive tests
	val, ok := jsonArr.Array()
	require.True(t, ok)
	require.Equal(t, arr, val)
	require.Equal(t, arr, jsonArr.Value())
	require.Equal(t, expectedUnwrapped, jsonArr.Unwrap())

	// Negative tests
	_, ok = jsonArr.Object()
	require.False(t, ok)
	_, ok = jsonArr.Number()
	require.False(t, ok)
	_, ok = jsonArr.String()
	require.False(t, ok)
	_, ok = jsonArr.Bool()
	require.False(t, ok)
	require.False(t, jsonArr.IsNull())
}

func TestJSONNumber_Methods_ShouldWorkAsExpected(t *testing.T) {
	num := newJSONNumber(2.5, nil)
	expected := 2.5

	// Positive tests
	val, ok := num.Number()
	require.True(t, ok)
	require.Equal(t, expected, val)
	require.Equal(t, expected, num.Value())
	require.Equal(t, expected, num.Unwrap())

	// Negative tests
	_, ok = num.Object()
	require.False(t, ok)
	_, ok = num.Array()
	require.False(t, ok)
	_, ok = num.String()
	require.False(t, ok)
	_, ok = num.Bool()
	require.False(t, ok)
	require.False(t, num.IsNull())
}

func TestJSONString_Methods_ShouldWorkAsExpected(t *testing.T) {
	str := newJSONString("value", nil)
	expected := "value"

	// Positive tests
	val, ok := str.String()
	require.True(t, ok)
	require.Equal(t, expected, val)
	require.Equal(t, expected, str.Value())
	require.Equal(t, expected, str.Unwrap())

	// Negative tests
	_, ok = str.Object()
	require.False(t, ok)
	_, ok = str.Array()
	require.False(t, ok)
	_, ok = str.Number()
	require.False(t, ok)
	_, ok = str.Bool()
	require.False(t, ok)
	require.False(t, str.IsNull())
}

func TestJSONBool_Methods_ShouldWorkAsExpected(t *testing.T) {
	b := newJSONBool(true, nil)
	expected := true

	// Positive tests
	val, ok := b.Bool()
	require.True(t, ok)
	require.Equal(t, expected, val)
	require.Equal(t, expected, b.Value())
	require.Equal(t, expected, b.Unwrap())

	// Negative tests
	_, ok = b.Object()
	require.False(t, ok)
	_, ok = b.Array()
	require.False(t, ok)
	_, ok = b.Number()
	require.False(t, ok)
	_, ok = b.String()
	require.False(t, ok)
	require.False(t, b.IsNull())
}

func TestJSONNull_Methods_ShouldWorkAsExpected(t *testing.T) {
	null := newJSONNull(nil)

	// Positive tests
	require.True(t, null.IsNull())
	require.Nil(t, null.Value())
	require.Nil(t, null.Unwrap())

	// Negative tests
	_, ok := null.Object()
	require.False(t, ok)
	_, ok = null.Array()
	require.False(t, ok)
	_, ok = null.Number()
	require.False(t, ok)
	_, ok = null.String()
	require.False(t, ok)
	_, ok = null.Bool()
	require.False(t, ok)
}

func TestNewJSONAndMarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		input        any
		expected     JSON
		expectedJSON string
		expectError  bool
	}{
		{
			name:         "Nil",
			input:        nil,
			expected:     newJSONNull(nil),
			expectedJSON: "null",
		},
		{
			name:         "FastJSON",
			input:        fastjson.MustParse(`{"key": "value"}`),
			expected:     newJSONObject(map[string]JSON{"key": newJSONString("value", nil)}, nil),
			expectedJSON: `{"key":"value"}`,
		},
		{
			name:         "Map",
			input:        map[string]any{"key": "value"},
			expected:     newJSONObject(map[string]JSON{"key": newJSONString("value", nil)}, nil),
			expectedJSON: `{"key":"value"}`,
		},
		{
			name:         "Bool",
			input:        true,
			expected:     newJSONBool(true, nil),
			expectedJSON: "true",
		},
		{
			name:         "String",
			input:        "str",
			expected:     newJSONString("str", nil),
			expectedJSON: `"str"`,
		},
		{
			name:         "Int8",
			input:        int8(42),
			expected:     newJSONNumber(42, nil),
			expectedJSON: "42",
		},
		{
			name:         "Int16",
			input:        int16(42),
			expected:     newJSONNumber(42, nil),
			expectedJSON: "42",
		},
		{
			name:         "Int32",
			input:        int32(42),
			expected:     newJSONNumber(42, nil),
			expectedJSON: "42",
		},
		{
			name:         "Int64",
			input:        int64(42),
			expected:     newJSONNumber(42, nil),
			expectedJSON: "42",
		},
		{
			name:         "Int",
			input:        42,
			expected:     newJSONNumber(42, nil),
			expectedJSON: "42",
		},
		{
			name:         "Uint8",
			input:        uint8(42),
			expected:     newJSONNumber(42, nil),
			expectedJSON: "42",
		},
		{
			name:         "Uint16",
			input:        uint16(42),
			expected:     newJSONNumber(42, nil),
			expectedJSON: "42",
		},
		{
			name:         "Uint32",
			input:        uint32(42),
			expected:     newJSONNumber(42, nil),
			expectedJSON: "42",
		},
		{
			name:         "Uint64",
			input:        uint64(42),
			expected:     newJSONNumber(42, nil),
			expectedJSON: "42",
		},
		{
			name:         "Uint",
			input:        uint(42),
			expected:     newJSONNumber(42, nil),
			expectedJSON: "42",
		},
		{
			name:         "Float32",
			input:        float32(2.5),
			expected:     newJSONNumber(2.5, nil),
			expectedJSON: "2.5",
		},
		{
			name:         "Float64",
			input:        float64(2.5),
			expected:     newJSONNumber(2.5, nil),
			expectedJSON: "2.5",
		},
		{
			name:         "BoolArray",
			input:        []bool{true, false},
			expected:     newJSONArray([]JSON{newJSONBool(true, nil), newJSONBool(false, nil)}, nil),
			expectedJSON: "[true,false]",
		},
		{
			name:  "StringArray",
			input: []string{"a", "b", "c"},
			expected: newJSONArray([]JSON{newJSONString("a", nil), newJSONString("b", nil),
				newJSONString("c", nil)}, nil),
			expectedJSON: `["a","b","c"]`,
		},
		{
			name:  "AnyArray",
			input: []any{"a", 1, true},
			expected: newJSONArray([]JSON{newJSONString("a", nil), newJSONNumber(1, nil),
				newJSONBool(true, nil)}, nil),
			expectedJSON: `["a",1,true]`,
		},
		{
			name:  "Int8Array",
			input: []int8{1, 2, 3},
			expected: newJSONArray([]JSON{newJSONNumber(1, nil), newJSONNumber(2, nil),
				newJSONNumber(3, nil)}, nil),
			expectedJSON: "[1,2,3]",
		},
		{
			name:  "Int16Array",
			input: []int16{1, 2, 3},
			expected: newJSONArray([]JSON{newJSONNumber(1, nil), newJSONNumber(2, nil),
				newJSONNumber(3, nil)}, nil),
			expectedJSON: "[1,2,3]",
		},
		{
			name:  "Int32Array",
			input: []int32{1, 2, 3},
			expected: newJSONArray([]JSON{newJSONNumber(1, nil), newJSONNumber(2, nil),
				newJSONNumber(3, nil)}, nil),
			expectedJSON: "[1,2,3]",
		},
		{
			name:  "Int64Array",
			input: []int64{1, 2, 3},
			expected: newJSONArray([]JSON{newJSONNumber(1, nil), newJSONNumber(2, nil),
				newJSONNumber(3, nil)}, nil),
			expectedJSON: "[1,2,3]",
		},
		{
			name:  "IntArray",
			input: []int{1, 2, 3},
			expected: newJSONArray([]JSON{newJSONNumber(1, nil), newJSONNumber(2, nil),
				newJSONNumber(3, nil)}, nil),
			expectedJSON: "[1,2,3]",
		},
		{
			name:  "Uint8Array",
			input: []uint8{1, 2, 3},
			expected: newJSONArray([]JSON{newJSONNumber(1, nil), newJSONNumber(2, nil),
				newJSONNumber(3, nil)}, nil),
			expectedJSON: "[1,2,3]",
		},
		{
			name:  "Uint16Array",
			input: []uint16{1, 2, 3},
			expected: newJSONArray([]JSON{newJSONNumber(1, nil), newJSONNumber(2, nil),
				newJSONNumber(3, nil)}, nil),
			expectedJSON: "[1,2,3]",
		},
		{
			name:  "Uint32Array",
			input: []uint32{1, 2, 3},
			expected: newJSONArray([]JSON{newJSONNumber(1, nil), newJSONNumber(2, nil),
				newJSONNumber(3, nil)}, nil),
			expectedJSON: "[1,2,3]",
		},
		{
			name:  "Uint64Array",
			input: []uint64{1, 2, 3},
			expected: newJSONArray([]JSON{newJSONNumber(1, nil), newJSONNumber(2, nil),
				newJSONNumber(3, nil)}, nil),
			expectedJSON: "[1,2,3]",
		},
		{
			name:  "UintArray",
			input: []uint{1, 2, 3},
			expected: newJSONArray([]JSON{newJSONNumber(1, nil), newJSONNumber(2, nil),
				newJSONNumber(3, nil)}, nil),
			expectedJSON: "[1,2,3]",
		},
		{
			name:  "Float32Array",
			input: []float32{1.0, 2.25, 3.5},
			expected: newJSONArray([]JSON{newJSONNumber(1.0, nil), newJSONNumber(2.25, nil),
				newJSONNumber(3.5, nil)}, nil),
			expectedJSON: "[1,2.25,3.5]",
		},
		{
			name:  "Float64Array",
			input: []float64{1.0, 2.25, 3.5},
			expected: newJSONArray([]JSON{newJSONNumber(1.0, nil), newJSONNumber(2.25, nil),
				newJSONNumber(3.5, nil)}, nil),
			expectedJSON: "[1,2.25,3.5]",
		},
		{
			name:        "AnyArrayWithInvalidElement",
			input:       []any{"valid", make(chan int)}, // channels can't be converted to JSON
			expectError: true,
		},
	}

	path := []string{"some", "path"}

	for _, tt := range tests {
		for _, withPath := range []bool{true, false} {
			t.Run(fmt.Sprintf("Test: %s, withPath: %v", tt.name, withPath), func(t *testing.T) {
				var result JSON
				var err error
				if withPath {
					result, err = NewJSONWithPath(tt.input, path)
				} else {
					result, err = NewJSON(tt.input)
				}
				if tt.expectError {
					require.Error(t, err, "Expected error, but got nil")
					return
				}
				require.NoError(t, err, "NewJSON failed with error %v", err)

				if withPath {
					traverseAndAssertPaths(t, result, path)
					require.Equal(t, result.Unwrap(), tt.expected.Unwrap())
					require.Equal(t, path, result.GetPath())
				} else {
					traverseAndAssertPaths(t, result, nil)
					require.Equal(t, result.Unwrap(), tt.expected.Unwrap())
					require.Empty(t, result.GetPath())
				}

				if !tt.expectError {
					jsonBytes, err := result.MarshalJSON()
					require.NoError(t, err, "MarshalJSON failed with error %v", err)
					require.Equal(t, tt.expectedJSON, string(jsonBytes))
				}
			})
		}
	}
}

func TestNewJSONFromMap_WithInvalidValue_ShouldFail(t *testing.T) {
	// Map with an invalid value (channel cannot be converted to JSON)
	input := map[string]any{
		"valid":   "value",
		"invalid": make(chan int),
	}

	_, err := NewJSONFromMap(input)
	require.Error(t, err)
}

func TestNewJSONFromMap_WithPaths(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected []struct {
			path  []string
			value any
		}
	}{
		{
			name: "flat object",
			input: map[string]any{
				"string": "value",
				"number": 42,
				"bool":   true,
				"null":   nil,
			},
			expected: []struct {
				path  []string
				value any
			}{
				{path: []string{"string"}, value: "value"},
				{path: []string{"number"}, value: float64(42)},
				{path: []string{"bool"}, value: true},
				{path: []string{"null"}, value: nil},
			},
		},
		{
			name: "nested object",
			input: map[string]any{
				"obj": map[string]any{
					"nested": "value",
					"deep": map[string]any{
						"number": 42,
					},
				},
				"arr": []any{
					"first",
					map[string]any{
						"inside_arr": true,
					},
					[]any{1, "nested"},
				},
			},
			expected: []struct {
				path  []string
				value any
			}{
				{path: []string{"obj", "nested"}, value: "value"},
				{path: []string{"obj", "deep", "number"}, value: float64(42)},
				{path: []string{"arr"}, value: "first"},
				{path: []string{"arr", "inside_arr"}, value: true},
				{path: []string{"arr"}, value: float64(1)},
				{path: []string{"arr"}, value: "nested"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			json, err := NewJSONFromMap(tt.input)
			require.NoError(t, err)

			traverseAndAssertPaths(t, json, nil)
		})
	}
}

func traverseAndAssertPaths(t *testing.T, j JSON, parentPath []string) {
	assert.Equal(t, parentPath, j.GetPath(), "Expected path %v, got %v", parentPath, j.GetPath())

	if obj, isObj := j.Object(); isObj {
		for k, v := range obj {
			newPath := append(parentPath, k)
			traverseAndAssertPaths(t, v, newPath)
		}
	}

	if arr, isArr := j.Array(); isArr {
		for _, v := range arr {
			traverseAndAssertPaths(t, v, parentPath)
		}
	}
}
