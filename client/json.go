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
	"encoding/json"
	"io"

	"github.com/valyala/fastjson"
	"golang.org/x/exp/constraints"
)

// JSON represents a JSON value that can be any valid JSON type: object, array, number, string, boolean, or null.
// It provides type-safe access to the underlying value through various accessor methods.
type JSON interface {
	// Array returns the value as a JSON array along with a boolean indicating if the value is an array.
	// Returns nil and false if the value is not an array.
	Array() ([]JSON, bool)

	// Object returns the value as a JSON object along with a boolean indicating if the value is an object.
	// Returns nil and false if the value is not an object.
	Object() (map[string]JSON, bool)

	// Number returns the value as a number along with a boolean indicating if the value is a number.
	// Returns 0 and false if the value is not a number.
	Number() (float64, bool)

	// String returns the value as a string along with a boolean indicating if the value is a string.
	// Returns empty string and false if the value is not a string.
	String() (string, bool)

	// Bool returns the value as a boolean along with a boolean indicating if the value is a boolean.
	// Returns false and false if the value is not a boolean.
	Bool() (bool, bool)

	// IsNull returns true if the value is null, false otherwise.
	IsNull() bool

	// Value returns the value that JSON represents.
	// The type will be one of: map[string]JSON, []JSON, float64, string, bool, or nil.
	Value() any

	// Unwrap returns the underlying value with all nested JSON values unwrapped.
	// For objects and arrays, this recursively unwraps all nested JSON values.
	Unwrap() any

	// Marshal writes the JSON value to the writer.
	// Returns an error if marshaling fails.
	Marshal(w io.Writer) error

	// MarshalJSON implements json.Marshaler interface.
	// Returns the JSON encoding of the value.
	MarshalJSON() ([]byte, error)
}

type jsonVoid struct{}

func (v jsonVoid) Object() (map[string]JSON, bool) {
	return nil, false
}

func (v jsonVoid) Array() ([]JSON, bool) {
	return nil, false
}

func (v jsonVoid) Number() (float64, bool) {
	return 0, false
}

func (v jsonVoid) String() (string, bool) {
	return "", false
}

func (v jsonVoid) Bool() (bool, bool) {
	return false, false
}

func (v jsonVoid) IsNull() bool {
	return false
}

type jsonBase[T any] struct {
	jsonVoid
	val T
}

func (v jsonBase[T]) Value() any {
	return v.val
}

func (v jsonBase[T]) Unwrap() any {
	return v.val
}

func (v jsonBase[T]) Marshal(w io.Writer) error {
	return json.NewEncoder(w).Encode(v.val)
}

func (v jsonBase[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.val)
}

type jsonObject struct {
	jsonBase[map[string]JSON]
}

var _ JSON = jsonObject{}

func (obj jsonObject) Object() (map[string]JSON, bool) {
	return obj.val, true
}

func (obj jsonObject) MarshalJSON() ([]byte, error) {
	return json.Marshal(obj.val)
}

func (obj jsonObject) Unwrap() any {
	result := make(map[string]any, len(obj.jsonBase.val))
	for k, v := range obj.val {
		result[k] = v.Unwrap()
	}
	return result
}

type jsonArray struct {
	jsonBase[[]JSON]
}

var _ JSON = jsonArray{}

func (arr jsonArray) Array() ([]JSON, bool) {
	return arr.val, true
}

func (arr jsonArray) MarshalJSON() ([]byte, error) {
	return json.Marshal(arr.val)
}

func (arr jsonArray) Unwrap() any {
	result := make([]any, len(arr.jsonBase.val))
	for i := range arr.val {
		result[i] = arr.val[i].Unwrap()
	}
	return result
}

type jsonNumber struct {
	jsonBase[float64]
}

var _ JSON = jsonNumber{}

func (n jsonNumber) Number() (float64, bool) {
	return n.val, true
}

func (n jsonNumber) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.val)
}

type jsonString struct {
	jsonBase[string]
}

var _ JSON = jsonString{}

func (s jsonString) String() (string, bool) {
	return s.val, true
}

func (s jsonString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.val)
}

type jsonBool struct {
	jsonBase[bool]
}

var _ JSON = jsonBool{}

func (b jsonBool) Bool() (bool, bool) {
	return b.val, true
}

func (b jsonBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.val)
}

type jsonNull struct {
	jsonVoid
}

var _ JSON = jsonNull{}

func (n jsonNull) IsNull() bool {
	return true
}

func (n jsonNull) Value() any {
	return nil
}

func (n jsonNull) Unwrap() any {
	return nil
}

func (n jsonNull) Marshal(w io.Writer) error {
	return json.NewEncoder(w).Encode(nil)
}

func (n jsonNull) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}

func newJSONObject(val map[string]JSON) JSON {
	return jsonObject{jsonBase[map[string]JSON]{val: val}}
}

func newJSONArray(val []JSON) JSON {
	return jsonArray{jsonBase[[]JSON]{val: val}}
}

func newJSONNumber(val float64) JSON {
	return jsonNumber{jsonBase[float64]{val: val}}
}

func newJSONString(val string) JSON {
	return jsonString{jsonBase[string]{val: val}}
}

func newJSONBool(val bool) JSON {
	return jsonBool{jsonBase[bool]{val: val}}
}

func newJSONNull() JSON {
	return jsonNull{}
}

// ParseJSONBytes parses the given JSON bytes into a JSON value.
// Returns error if the input is not valid JSON.
func ParseJSONBytes(data []byte) (JSON, error) {
	var p fastjson.Parser
	v, err := p.ParseBytes(data)
	if err != nil {
		return nil, err
	}
	return NewJSONFromFastJSON(v), nil
}

// ParseJSONString parses the given JSON string into a JSON value.
// Returns error if the input is not valid JSON.
func ParseJSONString(data string) (JSON, error) {
	var p fastjson.Parser
	v, err := p.Parse(data)
	if err != nil {
		return nil, err
	}
	return NewJSONFromFastJSON(v), nil
}

// NewJSON creates a JSON value from a Go value.
// The Go value must be one of:
// - nil (becomes JSON null)
// - *fastjson.Value
// - string
// - map[string]any
// - bool
// - numeric types (int8 through int64, uint8 through uint64, float32, float64)
// - slice of any above type
// - []any
// Returns error if the input cannot be converted to JSON.
func NewJSON(v any) (JSON, error) {
	if v == nil {
		return newJSONNull(), nil
	}
	switch val := v.(type) {
	case *fastjson.Value:
		return NewJSONFromFastJSON(val), nil
	case string:
		return newJSONString(val), nil
	case map[string]any:
		return NewJSONFromMap(val)
	case bool:
		return newJSONBool(val), nil
	case int8:
		return newJSONNumber(float64(val)), nil
	case int16:
		return newJSONNumber(float64(val)), nil
	case int32:
		return newJSONNumber(float64(val)), nil
	case int64:
		return newJSONNumber(float64(val)), nil
	case int:
		return newJSONNumber(float64(val)), nil
	case uint8:
		return newJSONNumber(float64(val)), nil
	case uint16:
		return newJSONNumber(float64(val)), nil
	case uint32:
		return newJSONNumber(float64(val)), nil
	case uint64:
		return newJSONNumber(float64(val)), nil
	case uint:
		return newJSONNumber(float64(val)), nil
	case float32:
		return newJSONNumber(float64(val)), nil
	case float64:
		return newJSONNumber(val), nil

	case []bool:
		return newJSONBoolArray(val), nil
	case []int8:
		return newJSONNumberArray(val), nil
	case []int16:
		return newJSONNumberArray(val), nil
	case []int32:
		return newJSONNumberArray(val), nil
	case []int64:
		return newJSONNumberArray(val), nil
	case []int:
		return newJSONNumberArray(val), nil
	case []uint8:
		return newJSONNumberArray(val), nil
	case []uint16:
		return newJSONNumberArray(val), nil
	case []uint32:
		return newJSONNumberArray(val), nil
	case []uint64:
		return newJSONNumberArray(val), nil
	case []uint:
		return newJSONNumberArray(val), nil
	case []float32:
		return newJSONNumberArray(val), nil
	case []float64:
		return newJSONNumberArray(val), nil
	case []string:
		return newJSONStringArray(val), nil

	case []any:
		arr := make([]JSON, 0)
		for _, item := range val {
			el, err := NewJSON(item)
			if err != nil {
				return nil, err
			}
			arr = append(arr, el)
		}
		return newJSONArray(arr), nil
	}

	return nil, NewErrInvalidJSONPayload(v)
}

func newJSONBoolArray(v []bool) JSON {
	arr := make([]JSON, 0)
	for _, item := range v {
		arr = append(arr, newJSONBool(item))
	}
	return newJSONArray(arr)
}

func newJSONNumberArray[T constraints.Integer | constraints.Float](v []T) JSON {
	arr := make([]JSON, 0)
	for _, item := range v {
		arr = append(arr, newJSONNumber(float64(item)))
	}
	return newJSONArray(arr)
}

func newJSONStringArray(v []string) JSON {
	arr := make([]JSON, 0)
	for _, item := range v {
		arr = append(arr, newJSONString(item))
	}
	return newJSONArray(arr)
}

// NewJSONFromFastJSON creates a JSON value from a fastjson.Value.
func NewJSONFromFastJSON(v *fastjson.Value) JSON {
	switch v.Type() {
	case fastjson.TypeObject:
		obj := make(map[string]JSON)
		v.GetObject().Visit(func(k []byte, v *fastjson.Value) {
			obj[string(k)] = NewJSONFromFastJSON(v)
		})
		return newJSONObject(obj)
	case fastjson.TypeArray:
		arr := make([]JSON, 0)
		for _, item := range v.GetArray() {
			arr = append(arr, NewJSONFromFastJSON(item))
		}
		return newJSONArray(arr)
	case fastjson.TypeNumber:
		return newJSONNumber(v.GetFloat64())
	case fastjson.TypeString:
		return newJSONString(string(v.GetStringBytes()))
	case fastjson.TypeTrue:
		return newJSONBool(true)
	case fastjson.TypeFalse:
		return newJSONBool(false)
	case fastjson.TypeNull:
		return newJSONNull()
	}
	return nil
}

// NewJSONFromMap creates a JSON object from a map[string]any.
// The map values must be valid Go values that can be converted to JSON.
// Returns error if any map value cannot be converted to JSON.
func NewJSONFromMap(data map[string]any) (JSON, error) {
	obj := make(map[string]JSON)
	for k, v := range data {
		jsonVal, err := NewJSON(v)
		if err != nil {
			return nil, err
		}
		obj[k] = jsonVal
	}
	return newJSONObject(obj), nil
}
