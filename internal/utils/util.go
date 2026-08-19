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
	"reflect"

	"github.com/sourcenetwork/immutable/enumerable"
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
func DecodeJSONVariables(data []byte) (map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var vars map[string]any
	if err := dec.Decode(&vars); err != nil {
		return nil, err
	}
	normalizeJSONNumbers(vars)
	return vars, nil
}

// normalizeJSONNumbers walks a value produced by a decoder configured with UseNumber,
// converting json.Number leaves into an int64 (if the number round-trips exactly) or a
// float64 (otherwise), so callers receive ordinary numeric types instead of json.Number.
func normalizeJSONNumbers(value any) any {
	switch value := value.(type) {
	case json.Number:
		if i, err := value.Int64(); err == nil {
			return i
		}
		if f, err := value.Float64(); err == nil {
			return f
		}
		return value.String()
	case map[string]any:
		for k, v := range value {
			value[k] = normalizeJSONNumbers(v)
		}
		return value
	case []any:
		for i, v := range value {
			value[i] = normalizeJSONNumbers(v)
		}
		return value
	default:
		return value
	}
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
