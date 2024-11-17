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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fastjson"
)

func TestNewJSONAndMarshal_WithValidInput_ShouldMarshal(t *testing.T) {
	tests := []struct {
		name     string
		fromFunc func(string) (JSON, error)
	}{
		{
			name:     "FromBytes",
			fromFunc: func(data string) (JSON, error) { return NewJSONFromBytes([]byte(data)) },
		},
		{
			name:     "FromString",
			fromFunc: NewJSONFromString,
		},
		{
			name: "FromFastJSON",
			fromFunc: func(data string) (JSON, error) {
				var p fastjson.Parser
				v, err := p.Parse(data)
				if err != nil {
					return nil, err
				}
				return NewJSONFromFastJSON(v)
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
		})
	}
}

func TestJSONCastMethods_ShouldCastCorrespondingAndRejectOthers(t *testing.T) {
	tests := []struct {
		name     string
		jsonObj  JSON
		expected any
	}{
		{
			name:     "Object",
			jsonObj:  newJSONObject(map[string]JSON{"key": newJSONString("value")}),
			expected: map[string]JSON{"key": newJSONString("value")},
		},
		{
			name:     "Array",
			jsonObj:  newJSONArray([]JSON{newJSONString("item1"), newJSONNumber(2)}),
			expected: []JSON{newJSONString("item1"), newJSONNumber(2)},
		},
		{
			name:     "Number",
			jsonObj:  newJSONNumber(2.5),
			expected: 2.5,
		},
		{
			name:     "String",
			jsonObj:  newJSONString("value"),
			expected: "value",
		},
		{
			name:     "Bool",
			jsonObj:  newJSONBool(true),
			expected: true,
		},
		{
			name:     "Null",
			jsonObj:  newJSONNull(),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.expected.(type) {
			case map[string]JSON:
				obj, ok := tt.jsonObj.Object()
				require.True(t, ok, "Expected JSON object, but got something else")
				require.True(t, equalMaps(obj, v), "Expected %v, got %v", v, obj)
				_, ok = tt.jsonObj.Array()
				require.False(t, ok, "Expected false for Array method")
				_, ok = tt.jsonObj.Number()
				require.False(t, ok, "Expected false for Number method")
				_, ok = tt.jsonObj.String()
				require.False(t, ok, "Expected false for String method")
				_, ok = tt.jsonObj.Bool()
				require.False(t, ok, "Expected false for Bool method")
				require.False(t, tt.jsonObj.IsNull(), "Expected false for IsNull method")
			case []JSON:
				arr, ok := tt.jsonObj.Array()
				require.True(t, ok, "Expected JSON array, but got something else")
				require.True(t, equalSlices(arr, v), "Expected %v, got %v", v, arr)
				_, ok = tt.jsonObj.Object()
				require.False(t, ok, "Expected false for Object method")
				_, ok = tt.jsonObj.Number()
				require.False(t, ok, "Expected false for Number method")
				_, ok = tt.jsonObj.String()
				require.False(t, ok, "Expected false for String method")
				_, ok = tt.jsonObj.Bool()
				require.False(t, ok, "Expected false for Bool method")
				require.False(t, tt.jsonObj.IsNull(), "Expected false for IsNull method")
			case float64:
				num, ok := tt.jsonObj.Number()
				require.True(t, ok, "Expected JSON number, but got something else")
				require.Equal(t, v, num, "Expected %v, got %v", v, num)
				_, ok = tt.jsonObj.Object()
				require.False(t, ok, "Expected false for Object method")
				_, ok = tt.jsonObj.Array()
				require.False(t, ok, "Expected false for Array method")
				_, ok = tt.jsonObj.String()
				require.False(t, ok, "Expected false for String method")
				_, ok = tt.jsonObj.Bool()
				require.False(t, ok, "Expected false for Bool method")
				require.False(t, tt.jsonObj.IsNull(), "Expected false for IsNull method")
			case string:
				str, ok := tt.jsonObj.String()
				require.True(t, ok, "Expected JSON string, but got something else")
				require.Equal(t, v, str, "Expected %v, got %v", v, str)
				_, ok = tt.jsonObj.Object()
				require.False(t, ok, "Expected false for Object method")
				_, ok = tt.jsonObj.Array()
				require.False(t, ok, "Expected false for Array method")
				_, ok = tt.jsonObj.Number()
				require.False(t, ok, "Expected false for Number method")
				_, ok = tt.jsonObj.Bool()
				require.False(t, ok, "Expected false for Bool method")
				require.False(t, tt.jsonObj.IsNull(), "Expected false for IsNull method")
			case bool:
				b, ok := tt.jsonObj.Bool()
				require.True(t, ok, "Expected JSON boolean, but got something else")
				require.Equal(t, v, b, "Expected %v, got %v", v, b)
				_, ok = tt.jsonObj.Object()
				require.False(t, ok, "Expected false for Object method")
				_, ok = tt.jsonObj.Array()
				require.False(t, ok, "Expected false for Array method")
				_, ok = tt.jsonObj.Number()
				require.False(t, ok, "Expected false for Number method")
				_, ok = tt.jsonObj.String()
				require.False(t, ok, "Expected false for String method")
				require.False(t, tt.jsonObj.IsNull(), "Expected false for IsNull method")
			default: // nil
				require.True(t, tt.jsonObj.IsNull(), "Expected JSON null, but got something else")
				_, ok := tt.jsonObj.Object()
				require.False(t, ok, "Expected false for Object method")
				_, ok = tt.jsonObj.Array()
				require.False(t, ok, "Expected false for Array method")
				_, ok = tt.jsonObj.Number()
				require.False(t, ok, "Expected false for Number method")
				_, ok = tt.jsonObj.String()
				require.False(t, ok, "Expected false for String method")
				_, ok = tt.jsonObj.Bool()
				require.False(t, ok, "Expected false for Bool method")
			}
		})
	}
}

func equalMaps(a, b map[string]JSON) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if !equalJSON(v, b[k]) {
			return false
		}
	}
	return true
}

func equalSlices(a, b []JSON) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !equalJSON(a[i], b[i]) {
			return false
		}
	}
	return true
}

func equalJSON(a, b JSON) bool {
	var bufA, bufB bytes.Buffer
	a.Marshal(&bufA)
	b.Marshal(&bufB)
	return bufA.String() == bufB.String()
}
