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
)

type JSON interface {
	Array() ([]JSON, bool)
	Object() (map[string]JSON, bool)
	Number() (float64, bool)
	String() (string, bool)
	Bool() (bool, bool)
	IsNull() bool
	Value() any
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

func (v jsonObject) Object() (map[string]JSON, bool) {
	return v.val, true
}

func (v jsonObject) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.val)
}

type jsonArray struct {
	jsonBase[[]JSON]
}

var _ JSON = jsonArray{}

func (v jsonArray) Array() ([]JSON, bool) {
	return v.val, true
}

func (v jsonArray) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.val)
}

type jsonNumber struct {
	jsonBase[float64]
}

var _ JSON = jsonNumber{}

func (v jsonNumber) Number() (float64, bool) {
	return v.val, true
}

func (v jsonNumber) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.val)
}

type jsonString struct {
	jsonBase[string]
}

var _ JSON = jsonString{}

func (v jsonString) String() (string, bool) {
	return v.val, true
}

func (v jsonString) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.val)
}

type jsonBool struct {
	jsonBase[bool]
}

var _ JSON = jsonBool{}

func (v jsonBool) Bool() (bool, bool) {
	return v.val, true
}

func (v jsonBool) Marshal(w io.Writer) error {
	return json.NewEncoder(w).Encode(v.val)
}

func (v jsonBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.val)
}

type jsonNull struct {
	jsonVoid
}

var _ JSON = jsonNull{}

func (v jsonNull) IsNull() bool {
	return true
}

func (v jsonNull) Value() any {
	return nil
}

func (v jsonNull) Marshal(w io.Writer) error {
	return json.NewEncoder(w).Encode(nil)
}

func (v jsonNull) MarshalJSON() ([]byte, error) {
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

func NewJSONFromBytes(data []byte) (JSON, error) {
	var p fastjson.Parser
	v, err := p.ParseBytes(data)
	if err != nil {
		return nil, err
	}
	return NewJSONFromFastJSON(v)
}

func NewJSONFromString(data string) (JSON, error) {
	var p fastjson.Parser
	v, err := p.Parse(data)
	if err != nil {
		return nil, err
	}
	return NewJSONFromFastJSON(v)
}

func NewJSONFromFastJSON(v *fastjson.Value) (JSON, error) {
	switch v.Type() {
	case fastjson.TypeObject:
		obj := make(map[string]JSON)
		var err error
		v.GetObject().Visit(func(k []byte, v *fastjson.Value) {
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
