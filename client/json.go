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
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/valyala/fastjson"
	"golang.org/x/exp/constraints"
)

// JSONPathPart represents a part of a JSON path.
// Json path can be either a property of an object or an index of an element in an array.
// For example, consider the following JSON:
//
//	{
//	  "custom": {
//	    "name": "John"
//	  },
//	  "0": {
//	    "val": 1
//	  },
//	  [
//	    {
//	      "val": 2
//		}
//	  ]
//	}
//
// The path to a top-level document is empty.
// The path to subtree { "name": "John" } can be described as "custom".
// The path to value "John" can be described as "custom.name".
// The paths to both values 1 and 2 can be described as "0.val":
// - for value 1 it's "0" property of the object and "val" property of the object
// - for value 2 it's "0" index of the array and "val" property of the object
// That's why we need to distinguish between properties and indices in the path.
type JSONPathPart struct {
	value any
}

// Property returns the property name if the part is a property, and a boolean indicating if the part is a property.
func (p JSONPathPart) Property() (string, bool) {
	v, ok := p.value.(string)
	return v, ok
}

// Index returns the index if the part is an index, and a boolean indicating if the part is an index.
func (p JSONPathPart) Index() (uint64, bool) {
	v, ok := p.value.(uint64)
	return v, ok
}

// JSONPath represents a path to a JSON value in a JSON tree.
type JSONPath []JSONPathPart

// Parts returns the parts of the JSON path.
func (p JSONPath) Parts() []JSONPathPart {
	return p
}

// AppendProperty appends a property part to the JSON path.
func (p JSONPath) AppendProperty(part string) JSONPath {
	return append(p, JSONPathPart{value: part})
}

// AppendIndex appends an index part to the JSON path.
func (p JSONPath) AppendIndex(part uint64) JSONPath {
	return append(p, JSONPathPart{value: part})
}

// String returns the string representation of the JSON path.
func (p JSONPath) String() string {
	var sb strings.Builder
	for i, part := range p {
		if prop, ok := part.Property(); ok {
			if i > 0 {
				sb.WriteByte('.')
			}
			sb.WriteString(prop)
		} else if index, ok := part.Index(); ok {
			sb.WriteByte('[')
			sb.WriteString(strconv.FormatUint(index, 10))
			sb.WriteByte(']')
		}
	}
	return sb.String()
}

// JSON represents a JSON value that can be any valid JSON type: object, array, number, string, boolean, or null.
// It can also represent a subtree of a JSON tree.
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

	// GetPath returns the path of the JSON value (or subtree) in the JSON tree.
	GetPath() JSONPath

	// visit calls the visitor function for the JSON value at the given path.
	visit(visitor JSONVisitor, path JSONPath, opts traverseJSONOptions) error
}

// MakeVoidJSON creates a JSON value that represents a void value with just a path.
// This is necessary purely for creating a json path prefix for storage queries.
// All other json values will be encoded with some value after the path which makes
// them unsuitable to build a path prefix.
func MakeVoidJSON(path JSONPath) JSON {
	return jsonBase[any]{path: path}
}

// TraverseJSON traverses a JSON value and calls the visitor function for each node.
// opts controls how the traversal is performed.
func TraverseJSON(j JSON, visitor JSONVisitor, opts ...traverseJSONOption) error {
	var options traverseJSONOptions
	for _, opt := range opts {
		opt(&options)
	}
	if shouldVisitPath(options.pathPrefix, nil) {
		return j.visit(visitor, JSONPath{}, options)
	}
	return nil
}

type traverseJSONOption func(*traverseJSONOptions)

// TraverseJSONWithPrefix returns a traverseJSONOption that sets the path prefix for the traversal.
// Only nodes with paths that start with the prefix will be visited.
func TraverseJSONWithPrefix(prefix JSONPath) traverseJSONOption {
	return func(opts *traverseJSONOptions) {
		opts.pathPrefix = prefix
	}
}

// TraverseJSONOnlyLeaves returns a traverseJSONOption that sets the traversal to visit only leaf nodes.
// Leaf nodes are nodes that do not have any children. This means that visitor function will not
// be called for objects or arrays and proceed with theirs children.
func TraverseJSONOnlyLeaves() traverseJSONOption {
	return func(opts *traverseJSONOptions) {
		opts.onlyLeaves = true
	}
}

// TraverseJSONVisitArrayElements returns a traverseJSONOption that sets the traversal to visit array elements.
// When this option is set, the visitor function will be called for each element of an array.
// If recurseElements is true, the visitor function will be called for each array element of type object or array.
func TraverseJSONVisitArrayElements(recurseElements bool) traverseJSONOption {
	return func(opts *traverseJSONOptions) {
		opts.visitArrayElements = true
		opts.recurseVisitedArrayElements = recurseElements
	}
}

// TraverseJSONWithArrayIndexInPath returns a traverseJSONOption that includes array indices in the path.
func TraverseJSONWithArrayIndexInPath() traverseJSONOption {
	return func(opts *traverseJSONOptions) {
		opts.includeArrayIndexInPath = true
	}
}

// JSONVisitor is a function that processes a JSON value at every node of the JSON tree.
// Returns an error if the processing fails.
type JSONVisitor func(value JSON) error

// traverseJSONOptions configures how the JSON tree is traversed.
type traverseJSONOptions struct {
	// onlyLeaves when true visits only leaf nodes (not objects or arrays)
	onlyLeaves bool
	// pathPrefix when set visits only paths that start with this prefix
	pathPrefix JSONPath
	// visitArrayElements when true visits array elements
	visitArrayElements bool
	// recurseVisitedArrayElements when true visits array elements recursively
	recurseVisitedArrayElements bool
	// includeArrayIndexInPath when true includes array indices in the path
	includeArrayIndexInPath bool
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

func (v jsonVoid) visit(visitor JSONVisitor, path JSONPath, opts traverseJSONOptions) error {
	return nil
}

type jsonBase[T any] struct {
	jsonVoid
	val  T
	path JSONPath
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

func (v jsonBase[T]) GetPath() JSONPath {
	return v.path
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

func (obj jsonObject) visit(visitor JSONVisitor, path JSONPath, opts traverseJSONOptions) error {
	obj.path = path
	if !opts.onlyLeaves && len(path) >= len(opts.pathPrefix) {
		if err := visitor(obj); err != nil {
			return err
		}
	}

	for k, v := range obj.val {
		newPath := path.AppendProperty(k)
		if !shouldVisitPath(opts.pathPrefix, newPath) {
			continue
		}

		if err := v.visit(visitor, newPath, opts); err != nil {
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

func (arr jsonArray) visit(visitor JSONVisitor, path JSONPath, opts traverseJSONOptions) error {
	arr.path = path
	if !opts.onlyLeaves {
		if err := visitor(arr); err != nil {
			return err
		}
	}

	if opts.visitArrayElements {
		for i := range arr.val {
			if !opts.recurseVisitedArrayElements && isCompositeJSON(arr.val[i]) {
				continue
			}
			var newPath JSONPath
			if opts.includeArrayIndexInPath {
				newPath = path.AppendIndex(uint64(i))
			} else {
				newPath = path
			}
			if !shouldVisitPath(opts.pathPrefix, newPath) {
				continue
			}

			if err := arr.val[i].visit(visitor, newPath, opts); err != nil {
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

func (n jsonNumber) visit(visitor JSONVisitor, path JSONPath, opts traverseJSONOptions) error {
	n.path = path
	return visitor(n)
}

type jsonString struct {
	jsonBase[string]
}

var _ JSON = jsonString{}

func (s jsonString) String() (string, bool) {
	return s.val, true
}

func (n jsonString) visit(visitor JSONVisitor, path JSONPath, opts traverseJSONOptions) error {
	n.path = path
	return visitor(n)
}

type jsonBool struct {
	jsonBase[bool]
}

var _ JSON = jsonBool{}

func (b jsonBool) Bool() (bool, bool) {
	return b.val, true
}

func (n jsonBool) visit(visitor JSONVisitor, path JSONPath, opts traverseJSONOptions) error {
	n.path = path
	return visitor(n)
}

type jsonNull struct {
	jsonBase[any]
}

var _ JSON = jsonNull{}

func (n jsonNull) IsNull() bool {
	return true
}

func (n jsonNull) visit(visitor JSONVisitor, path JSONPath, opts traverseJSONOptions) error {
	n.path = path
	return visitor(n)
}

func newJSONObject(val map[string]JSON, path JSONPath) jsonObject {
	return jsonObject{jsonBase[map[string]JSON]{val: val, path: path}}
}

func newJSONArray(val []JSON, path JSONPath) jsonArray {
	return jsonArray{jsonBase[[]JSON]{val: val, path: path}}
}

func newJSONNumber(val float64, path JSONPath) jsonNumber {
	return jsonNumber{jsonBase[float64]{val: val, path: path}}
}

func newJSONString(val string, path JSONPath) jsonString {
	return jsonString{jsonBase[string]{val: val, path: path}}
}

func newJSONBool(val bool, path JSONPath) jsonBool {
	return jsonBool{jsonBase[bool]{val: val, path: path}}
}

func newJSONNull(path JSONPath) jsonNull {
	return jsonNull{jsonBase[any]{path: path}}
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
	return newJSON(v, nil)
}

// NewJSONWithPath creates a JSON value from a Go value with stored path to the value.
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
func NewJSONWithPath(v any, path JSONPath) (JSON, error) {
	return newJSON(v, path)
}

// newJSON is an internal function that creates a new JSON value with parent and property name
func newJSON(v any, path JSONPath) (JSON, error) {
	if v == nil {
		return newJSONNull(path), nil
	} else {
		switch val := v.(type) {
		case *fastjson.Value:
			return newJSONFromFastJSON(val, path), nil
		case string:
			return newJSONString(val, path), nil
		case map[string]any:
			return newJSONFromMap(val, path)
		case bool:
			return newJSONBool(val, path), nil
		case int8:
			return newJSONNumber(float64(val), path), nil
		case int16:
			return newJSONNumber(float64(val), path), nil
		case int32:
			return newJSONNumber(float64(val), path), nil
		case int64:
			return newJSONNumber(float64(val), path), nil
		case int:
			return newJSONNumber(float64(val), path), nil
		case uint8:
			return newJSONNumber(float64(val), path), nil
		case uint16:
			return newJSONNumber(float64(val), path), nil
		case uint32:
			return newJSONNumber(float64(val), path), nil
		case uint64:
			return newJSONNumber(float64(val), path), nil
		case uint:
			return newJSONNumber(float64(val), path), nil
		case float32:
			return newJSONNumber(float64(val), path), nil
		case float64:
			return newJSONNumber(val, path), nil

		case []bool:
			return newJSONBoolArray(val, path), nil
		case []int8:
			return newJSONNumberArray(val, path), nil
		case []int16:
			return newJSONNumberArray(val, path), nil
		case []int32:
			return newJSONNumberArray(val, path), nil
		case []int64:
			return newJSONNumberArray(val, path), nil
		case []int:
			return newJSONNumberArray(val, path), nil
		case []uint8:
			return newJSONNumberArray(val, path), nil
		case []uint16:
			return newJSONNumberArray(val, path), nil
		case []uint32:
			return newJSONNumberArray(val, path), nil
		case []uint64:
			return newJSONNumberArray(val, path), nil
		case []uint:
			return newJSONNumberArray(val, path), nil
		case []float32:
			return newJSONNumberArray(val, path), nil
		case []float64:
			return newJSONNumberArray(val, path), nil
		case []string:
			return newJSONStringArray(val, path), nil
		case []any:
			return newJsonArrayFromAnyArray(val, path)
		}
	}

	return nil, NewErrInvalidJSONPayload(v)
}

func newJsonArrayFromAnyArray(arr []any, path JSONPath) (JSON, error) {
	result := make([]JSON, len(arr))
	for i := range arr {
		jsonVal, err := newJSON(arr[i], path.AppendIndex(uint64(i)))
		if err != nil {
			return nil, err
		}
		result[i] = jsonVal
	}
	return newJSONArray(result, path), nil
}

func newJSONBoolArray(v []bool, path JSONPath) JSON {
	arr := make([]JSON, len(v))
	for i := range v {
		arr[i] = newJSONBool(v[i], path.AppendIndex(uint64(i)))
	}
	return newJSONArray(arr, path)
}

func newJSONNumberArray[T constraints.Integer | constraints.Float](v []T, path JSONPath) JSON {
	arr := make([]JSON, len(v))
	for i := range v {
		arr[i] = newJSONNumber(float64(v[i]), path.AppendIndex(uint64(i)))
	}
	return newJSONArray(arr, path)
}

func newJSONStringArray(v []string, path JSONPath) JSON {
	arr := make([]JSON, len(v))
	for i := range v {
		arr[i] = newJSONString(v[i], path.AppendIndex(uint64(i)))
	}
	return newJSONArray(arr, path)
}

// newJSONFromFastJSON is an internal function that creates a new JSON value with parent and property name
func newJSONFromFastJSON(v *fastjson.Value, path JSONPath) JSON {
	switch v.Type() {
	case fastjson.TypeObject:
		fastObj := v.GetObject()
		obj := make(map[string]JSON, fastObj.Len())
		fastObj.Visit(func(k []byte, v *fastjson.Value) {
			key := string(k)
			obj[key] = newJSONFromFastJSON(v, path.AppendProperty(key))
		})
		return newJSONObject(obj, path)
	case fastjson.TypeArray:
		fastArr := v.GetArray()
		arr := make([]JSON, len(fastArr))
		for i := range fastArr {
			arr[i] = newJSONFromFastJSON(fastArr[i], path.AppendIndex(uint64(i)))
		}
		return newJSONArray(arr, path)
	case fastjson.TypeNumber:
		return newJSONNumber(v.GetFloat64(), path)
	case fastjson.TypeString:
		return newJSONString(string(v.GetStringBytes()), path)
	case fastjson.TypeTrue:
		return newJSONBool(true, path)
	case fastjson.TypeFalse:
		return newJSONBool(false, path)
	case fastjson.TypeNull:
		return newJSONNull(path)
	}
	return nil
}

// NewJSONFromFastJSON creates a JSON value from a fastjson.Value.
func NewJSONFromFastJSON(v *fastjson.Value) JSON {
	return newJSONFromFastJSON(v, nil)
}

// NewJSONFromMap creates a JSON object from a map[string]any.
// The map values must be valid Go values that can be converted to JSON.
// Returns error if any map value cannot be converted to JSON.
func NewJSONFromMap(data map[string]any) (JSON, error) {
	return newJSONFromMap(data, nil)
}

func newJSONFromMap(data map[string]any, path JSONPath) (JSON, error) {
	obj := make(map[string]JSON, len(data))
	for k, v := range data {
		jsonVal, err := newJSON(v, path.AppendProperty(k))
		if err != nil {
			return nil, err
		}
		obj[k] = jsonVal
	}
	return newJSONObject(obj, path), nil
}

func shouldVisitPath(prefix, path JSONPath) bool {
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

func isCompositeJSON(v JSON) bool {
	_, isObject := v.Object()
	if isObject {
		return true
	}
	_, isArray := v.Array()
	return isArray
}
