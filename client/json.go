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
	"strconv"

	"github.com/valyala/fastjson"
	"golang.org/x/exp/constraints"
)

// JSON represents a JSON value that can be any valid JSON type: object, array, number, string, boolean, or null.
// It provides type-safe access to the underlying value through various accessor methods.
type JSON interface {
	json.Marshaler
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

	// accept calls the visitor function for the JSON value at the given path.
	accept(visitor JSONVisitor, path []string, opts traverseJSONOptions) error
}

// TraverseJSON traverses a JSON value and calls the visitor function for each node.
// opts controls how the traversal is performed.
func TraverseJSON(j JSON, visitor JSONVisitor, opts ...traverseJSONOption) error {
	var options traverseJSONOptions
	for _, opt := range opts {
		opt(&options)
	}
	if shouldVisitPath(options.PathPrefix, nil) {
		return j.accept(visitor, []string{}, options)
	}
	return nil
}

type traverseJSONOption func(*traverseJSONOptions)

// TraverseJSONWithPrefix returns a traverseJSONOption that sets the path prefix for the traversal.
// Only nodes with paths that start with the prefix will be visited.
func TraverseJSONWithPrefix(prefix []string) traverseJSONOption {
	return func(opts *traverseJSONOptions) {
		opts.PathPrefix = prefix
	}
}

// TraverseJSONOnlyLeaves returns a traverseJSONOption that sets the traversal to visit only leaf nodes.
// Leaf nodes are nodes that do not have any children. This means that visitor function will not
// be called for objects or arrays and proceed with theirs children.
func TraverseJSONOnlyLeaves() traverseJSONOption {
	return func(opts *traverseJSONOptions) {
		opts.OnlyLeaves = true
	}
}

// TraverseJSONVisitArrayElements returns a traverseJSONOption that sets the traversal to visit array elements.
// When this option is set, the visitor function will be called for each element of an array.
func TraverseJSONVisitArrayElements() traverseJSONOption {
	return func(opts *traverseJSONOptions) {
		opts.VisitArrayElements = true
	}
}

// JSONVisitor is a function that processes a JSON value at a given path.
// path represents the location of the value in the JSON tree.
// Returns an error if the processing fails.
type JSONVisitor func(path []string, value JSON) error

// traverseJSONOptions configures how the JSON tree is traversed.
type traverseJSONOptions struct {
	// OnlyLeaves when true visits only leaf nodes (not objects or arrays)
	OnlyLeaves bool
	// PathPrefix when set visits only paths that start with this prefix
	PathPrefix []string
	// VisitArrayElements when true visits array elements
	VisitArrayElements bool
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

func (n jsonBase[T]) accept(visitor JSONVisitor, path []string, opts traverseJSONOptions) error {
	if shouldVisitPath(opts.PathPrefix, path) {
		return visitor(path, n)
	}
	return nil
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

func (obj jsonObject) accept(visitor JSONVisitor, path []string, opts traverseJSONOptions) error {
	if !opts.OnlyLeaves && len(path) >= len(opts.PathPrefix) {
		if err := visitor(path, obj); err != nil {
			return err
		}
	}

	for k, v := range obj.val {
		newPath := append(path, k)
		if !shouldVisitPath(opts.PathPrefix, newPath) {
			continue
		}

		if err := v.accept(visitor, newPath, opts); err != nil {
			return err
		}
	}
	return nil
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

func (arr jsonArray) accept(visitor JSONVisitor, path []string, opts traverseJSONOptions) error {
	if !opts.OnlyLeaves {
		if err := visitor(path, arr); err != nil {
			return err
		}
	}

	if opts.VisitArrayElements {
		for i := range arr.val {
			newPath := append(path, strconv.Itoa(i))
			if !shouldVisitPath(opts.PathPrefix, newPath) {
				continue
			}

			if err := arr.val[i].accept(visitor, newPath, opts); err != nil {
				return err
			}
		}
	}
	return nil
}

type jsonNumber struct {
	jsonBase[float64]
}

var _ JSON = jsonNumber{}

func (n jsonNumber) Number() (float64, bool) {
	return n.val, true
}

type jsonString struct {
	jsonBase[string]
}

var _ JSON = jsonString{}

func (s jsonString) String() (string, bool) {
	return s.val, true
}

type jsonBool struct {
	jsonBase[bool]
}

var _ JSON = jsonBool{}

func (b jsonBool) Bool() (bool, bool) {
	return b.val, true
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

func (n jsonNull) accept(visitor JSONVisitor, path []string, opts traverseJSONOptions) error {
	return visitor(path, n)
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
	// we could have called ParseJSONBytes([]byte(data), but this would copy the string to a byte slice.
	// fastjson.Parser.ParseBytes casts the bytes slice to a string internally, so we can avoid the extra copy.
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
		return newJsonArrayFromAnyArray(val)
	}

	return nil, NewErrInvalidJSONPayload(v)
}

func newJsonArrayFromAnyArray(arr []any) (JSON, error) {
	result := make([]JSON, len(arr))
	for i := range arr {
		jsonVal, err := NewJSON(arr[i])
		if err != nil {
			return nil, err
		}
		result[i] = jsonVal
	}
	return newJSONArray(result), nil
}

func newJSONBoolArray(v []bool) JSON {
	arr := make([]JSON, len(v))
	for i := range v {
		arr[i] = newJSONBool(v[i])
	}
	return newJSONArray(arr)
}

func newJSONNumberArray[T constraints.Integer | constraints.Float](v []T) JSON {
	arr := make([]JSON, len(v))
	for i := range v {
		arr[i] = newJSONNumber(float64(v[i]))
	}
	return newJSONArray(arr)
}

func newJSONStringArray(v []string) JSON {
	arr := make([]JSON, len(v))
	for i := range v {
		arr[i] = newJSONString(v[i])
	}
	return newJSONArray(arr)
}

// NewJSONFromFastJSON creates a JSON value from a fastjson.Value.
func NewJSONFromFastJSON(v *fastjson.Value) JSON {
	switch v.Type() {
	case fastjson.TypeObject:
		fastObj := v.GetObject()
		obj := make(map[string]JSON, fastObj.Len())
		fastObj.Visit(func(k []byte, v *fastjson.Value) {
			obj[string(k)] = NewJSONFromFastJSON(v)
		})
		return newJSONObject(obj)
	case fastjson.TypeArray:
		fastArr := v.GetArray()
		arr := make([]JSON, len(fastArr))
		for i := range fastArr {
			arr[i] = NewJSONFromFastJSON(fastArr[i])
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
	obj := make(map[string]JSON, len(data))
	for k, v := range data {
		jsonVal, err := NewJSON(v)
		if err != nil {
			return nil, err
		}
		obj[k] = jsonVal
	}
	return newJSONObject(obj), nil
}

func shouldVisitPath(prefix, path []string) bool {
	if len(prefix) == 0 {
		return true
	}
	for i := range prefix {
		if len(path) <= i {
			return true
		}
		if prefix[i] != path[i] {
			return false
		}
	}
	return true
}
