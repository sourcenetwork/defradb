// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client/request"
)

// NewOptions merges multiple option builders into a single options struct.
// It applies all functional options from all builders in the order they are provided.
// Returns nil if no valid options are provided.
//
// This follows the MongoDB Go driver pattern for option merging.
// Option builders implement enumerable.Enumerable, allowing iteration via Next()/Value().
//
// Example usage:
//
//	opts := options.NewOptions(
//	    options.GetCollections().SetIdentity(id),
//	    options.GetCollections().SetVersionID(vid),
//	)
func NewOptions[T any](opts ...enumerable.Enumerable[func(*T)]) *T {
	args := new(T)
	ApplyOptions(args, opts...)
	return args
}

// DecodeJSONVariables decodes a JSON-encoded GraphQL variables object without losing
// integer precision. The standard decoder represents every JSON number as a float64,
// which silently rounds any integer above 2^53. This instead reconstructs each number
// as an int64 where the value round-trips exactly, falling back to float64 otherwise.
// Returns an error if a number is out of range for both int64 and float64 (root or
// nested), or if data contains anything beyond a single JSON value.
func DecodeJSONVariables(data []byte) (map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var vars map[string]any
	if err := dec.Decode(&vars); err != nil {
		return nil, err
	}
	if dec.More() {
		return nil, fmt.Errorf("unexpected data after JSON variables object")
	}

	if _, err := normalizeJSONNumbers(vars); err != nil {
		return nil, err
	}
	return vars, nil
}

// normalizeJSONNumbers walks a value produced by a decoder configured with UseNumber,
// converting json.Number leaves into an int64 (if the number round-trips exactly) or a
// float64 (otherwise), so callers receive ordinary numeric types instead of json.Number.
// Returns an error if a number is out of range for both.
func normalizeJSONNumbers(value any) (any, error) {
	switch value := value.(type) {
	case json.Number:
		if i, err := value.Int64(); err == nil {
			return i, nil
		}
		if f, err := value.Float64(); err == nil {
			return f, nil
		}
		return nil, fmt.Errorf("json number %q is out of range", value.String())
	case map[string]any:
		for k, v := range value {
			normalized, err := normalizeJSONNumbers(v)
			if err != nil {
				return nil, err
			}
			value[k] = normalized
		}
		return value, nil
	case []any:
		for i, v := range value {
			normalized, err := normalizeJSONNumbers(v)
			if err != nil {
				return nil, err
			}
			value[i] = normalized
		}
		return value, nil
	default:
		return value, nil
	}
}

// DecodeJSONFilter decodes a JSON-encoded document filter without losing integer
// precision (see DecodeJSONVariables). Unlike DecodeJSONVariables, the decoded value
// is not assumed to be an object: data may be a filter expression (a string), a map 
// of filter conditions (an object), or null.
func DecodeJSONFilter(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	if dec.More() {
		return nil, fmt.Errorf("unexpected data after JSON filter value")
	}

	return normalizeJSONNumbers(value)
}

// NormalizeFilterForJSON prepares a document filter for JSON transport (e.g. across the C
// bindings boundary, or an HTTP/CLI request body). immutable.Option[request.Filter] does
// not round-trip through plain JSON: Some marshals to {"Conditions": ...}, indistinguishable
// on decode from a raw filter conditions map with a literal "Conditions" field, and None
// marshals to null, which decodes to an untyped nil that no filter type switch accepts. This
// unwraps Some to its Conditions map, and None to an empty conditions map (which still means
// "match all documents" once decoded). Any other filter value is returned unchanged.
func NormalizeFilterForJSON(filter any) any {
	if opt, ok := filter.(immutable.Option[request.Filter]); ok {
		if !opt.HasValue() {
			return map[string]any{}
		}
		return opt.Value().Conditions
	}
	return filter
}

// ApplyOptions applies all functional options onto the given target.
func ApplyOptions[T any](target *T, opts ...enumerable.Enumerable[func(*T)]) {
	for _, opt := range opts {
		if opt == nil || reflect.ValueOf(opt).IsNil() {
			continue
		}
		for {
			hasNext, err := opt.Next()
			if err != nil || !hasNext {
				break
			}
			setArgs, err := opt.Value()
			if err != nil {
				break
			}
			if setArgs != nil {
				setArgs(target)
			}
		}
		opt.Reset()
	}
}
