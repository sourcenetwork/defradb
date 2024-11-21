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

type JSON interface {
	Array() ([]JSON, bool)
	Object() (map[string]JSON, bool)
	Number() (float64, bool)
	String() (string, bool)
	Bool() (bool, bool)
	IsNull() bool
	Value() any
	Unwrap() any
	Marshal(w io.Writer) error
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

func (b jsonBool) Marshal(w io.Writer) error {
	return json.NewEncoder(w).Encode(b.val)
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

func ParseJSONBytes(data []byte) (JSON, error) {
	var p fastjson.Parser
	v, err := p.ParseBytes(data)
	if err != nil {
		return nil, err
	}
	return NewJSONFromFastJSON(v)
}

func ParseJSONString(data string) (JSON, error) {
	var p fastjson.Parser
	v, err := p.Parse(data)
	if err != nil {
		return nil, err
	}
	return NewJSONFromFastJSON(v)
}

func NewJSON(v any) (JSON, error) {
	if v == nil {
		return newJSONNull(), nil
	}
	switch val := v.(type) {
	case *fastjson.Value:
		return NewJSONFromFastJSON(val)
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

func newJSONBoolArray(v any) JSON {
	arr := make([]JSON, 0)
	for _, item := range v.([]bool) {
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

func NewJSONFromFastJSON(v *fastjson.Value) (JSON, error) {
	switch v.Type() {
	case fastjson.TypeObject:
		obj := make(map[string]JSON)
		var err error
		v.GetObject().Visit(func(k []byte, v *fastjson.Value) {
			if err != nil {
				return
			}
			val, newErr := NewJSONFromFastJSON(v)
			if newErr != nil {
				err = newErr
				return
			}
			obj[string(k)] = val
		})
		if err != nil {
			return nil, err
		}
		return newJSONObject(obj), nil
	case fastjson.TypeArray:
		arr := make([]JSON, 0)
		for _, item := range v.GetArray() {
			el, err := NewJSONFromFastJSON(item)
			if err != nil {
				return nil, err
			}
			arr = append(arr, el)
		}
		return newJSONArray(arr), nil
	case fastjson.TypeNumber:
		return newJSONNumber(v.GetFloat64()), nil
	case fastjson.TypeString:
		return newJSONString(string(v.GetStringBytes())), nil
	case fastjson.TypeTrue:
		return newJSONBool(true), nil
	case fastjson.TypeFalse:
		return newJSONBool(false), nil
	case fastjson.TypeNull:
		return newJSONNull(), nil
	default:
		return nil, NewErrInvalidJSONPayload(v)
	}
}

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
