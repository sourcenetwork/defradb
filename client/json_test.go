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
	"bytes"
	"encoding/json"
	"strings"
	"testing"

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
		"key": newJSONString("value"),
		"nested": newJSONObject(map[string]JSON{
			"inner": newJSONNumber(42),
			"array": newJSONArray([]JSON{newJSONString("test"), newJSONBool(true)}),
		}),
	}
	obj := newJSONObject(m)
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
		newJSONString("item1"),
		newJSONObject(map[string]JSON{
			"key": newJSONString("value"),
			"num": newJSONNumber(42),
		}),
		newJSONNumber(2),
	}
	jsonArr := newJSONArray(arr)
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
	num := newJSONNumber(2.5)
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
	str := newJSONString("value")
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
	b := newJSONBool(true)
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
	null := newJSONNull()

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
			expected:     newJSONNull(),
			expectedJSON: "null",
		},
		{
			name:         "FastJSON",
			input:        fastjson.MustParse(`{"key": "value"}`),
			expected:     newJSONObject(map[string]JSON{"key": newJSONString("value")}),
			expectedJSON: `{"key":"value"}`,
		},
		{
			name:         "Map",
			input:        map[string]any{"key": "value"},
			expected:     newJSONObject(map[string]JSON{"key": newJSONString("value")}),
			expectedJSON: `{"key":"value"}`,
		},
		{
			name:         "Bool",
			input:        true,
			expected:     newJSONBool(true),
			expectedJSON: "true",
		},
		{
			name:         "String",
			input:        "str",
			expected:     newJSONString("str"),
			expectedJSON: `"str"`,
		},
		{
			name:         "Int8",
			input:        int8(42),
			expected:     newJSONNumber(42),
			expectedJSON: "42",
		},
		{
			name:         "Int16",
			input:        int16(42),
			expected:     newJSONNumber(42),
			expectedJSON: "42",
		},
		{
			name:         "Int32",
			input:        int32(42),
			expected:     newJSONNumber(42),
			expectedJSON: "42",
		},
		{
			name:         "Int64",
			input:        int64(42),
			expected:     newJSONNumber(42),
			expectedJSON: "42",
		},
		{
			name:         "Int",
			input:        42,
			expected:     newJSONNumber(42),
			expectedJSON: "42",
		},
		{
			name:         "Uint8",
			input:        uint8(42),
			expected:     newJSONNumber(42),
			expectedJSON: "42",
		},
		{
			name:         "Uint16",
			input:        uint16(42),
			expected:     newJSONNumber(42),
			expectedJSON: "42",
		},
		{
			name:         "Uint32",
			input:        uint32(42),
			expected:     newJSONNumber(42),
			expectedJSON: "42",
		},
		{
			name:         "Uint64",
			input:        uint64(42),
			expected:     newJSONNumber(42),
			expectedJSON: "42",
		},
		{
			name:         "Uint",
			input:        uint(42),
			expected:     newJSONNumber(42),
			expectedJSON: "42",
		},
		{
			name:         "Float32",
			input:        float32(2.5),
			expected:     newJSONNumber(2.5),
			expectedJSON: "2.5",
		},
		{
			name:         "Float64",
			input:        float64(2.5),
			expected:     newJSONNumber(2.5),
			expectedJSON: "2.5",
		},
		{
			name:         "BoolArray",
			input:        []bool{true, false},
			expected:     newJSONArray([]JSON{newJSONBool(true), newJSONBool(false)}),
			expectedJSON: "[true,false]",
		},
		{
			name:         "StringArray",
			input:        []string{"a", "b", "c"},
			expected:     newJSONArray([]JSON{newJSONString("a"), newJSONString("b"), newJSONString("c")}),
			expectedJSON: `["a","b","c"]`,
		},
		{
			name:         "AnyArray",
			input:        []any{"a", 1, true},
			expected:     newJSONArray([]JSON{newJSONString("a"), newJSONNumber(1), newJSONBool(true)}),
			expectedJSON: `["a",1,true]`,
		},
		{
			name:         "Int8Array",
			input:        []int8{1, 2, 3},
			expected:     newJSONArray([]JSON{newJSONNumber(1), newJSONNumber(2), newJSONNumber(3)}),
			expectedJSON: "[1,2,3]",
		},
		{
			name:         "Int16Array",
			input:        []int16{1, 2, 3},
			expected:     newJSONArray([]JSON{newJSONNumber(1), newJSONNumber(2), newJSONNumber(3)}),
			expectedJSON: "[1,2,3]",
		},
		{
			name:         "Int32Array",
			input:        []int32{1, 2, 3},
			expected:     newJSONArray([]JSON{newJSONNumber(1), newJSONNumber(2), newJSONNumber(3)}),
			expectedJSON: "[1,2,3]",
		},
		{
			name:         "Int64Array",
			input:        []int64{1, 2, 3},
			expected:     newJSONArray([]JSON{newJSONNumber(1), newJSONNumber(2), newJSONNumber(3)}),
			expectedJSON: "[1,2,3]",
		},
		{
			name:         "IntArray",
			input:        []int{1, 2, 3},
			expected:     newJSONArray([]JSON{newJSONNumber(1), newJSONNumber(2), newJSONNumber(3)}),
			expectedJSON: "[1,2,3]",
		},
		{
			name:         "Uint8Array",
			input:        []uint8{1, 2, 3},
			expected:     newJSONArray([]JSON{newJSONNumber(1), newJSONNumber(2), newJSONNumber(3)}),
			expectedJSON: "[1,2,3]",
		},
		{
			name:         "Uint16Array",
			input:        []uint16{1, 2, 3},
			expected:     newJSONArray([]JSON{newJSONNumber(1), newJSONNumber(2), newJSONNumber(3)}),
			expectedJSON: "[1,2,3]",
		},
		{
			name:         "Uint32Array",
			input:        []uint32{1, 2, 3},
			expected:     newJSONArray([]JSON{newJSONNumber(1), newJSONNumber(2), newJSONNumber(3)}),
			expectedJSON: "[1,2,3]",
		},
		{
			name:         "Uint64Array",
			input:        []uint64{1, 2, 3},
			expected:     newJSONArray([]JSON{newJSONNumber(1), newJSONNumber(2), newJSONNumber(3)}),
			expectedJSON: "[1,2,3]",
		},
		{
			name:         "UintArray",
			input:        []uint{1, 2, 3},
			expected:     newJSONArray([]JSON{newJSONNumber(1), newJSONNumber(2), newJSONNumber(3)}),
			expectedJSON: "[1,2,3]",
		},
		{
			name:         "Float32Array",
			input:        []float32{1.0, 2.25, 3.5},
			expected:     newJSONArray([]JSON{newJSONNumber(1.0), newJSONNumber(2.25), newJSONNumber(3.5)}),
			expectedJSON: "[1,2.25,3.5]",
		},
		{
			name:         "Float64Array",
			input:        []float64{1.0, 2.25, 3.5},
			expected:     newJSONArray([]JSON{newJSONNumber(1.0), newJSONNumber(2.25), newJSONNumber(3.5)}),
			expectedJSON: "[1,2.25,3.5]",
		},
		{
			name:        "AnyArrayWithInvalidElement",
			input:       []any{"valid", make(chan int)}, // channels can't be converted to JSON
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewJSON(tt.input)
			if tt.expectError {
				require.Error(t, err, "Expected error, but got nil")
				return
			}
			require.NoError(t, err, "NewJSON failed with error %v", err)
			require.Equal(t, result, tt.expected)

			if !tt.expectError {
				jsonBytes, err := result.MarshalJSON()
				require.NoError(t, err, "MarshalJSON failed with error %v", err)
				require.Equal(t, tt.expectedJSON, string(jsonBytes))
			}
		})
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
