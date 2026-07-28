// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package types

import (
	"encoding/hex"
	"regexp"
	"strconv"

	"github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"

	"github.com/sourcenetwork/defradb/client/request"
)

// BlobPattern is a regex for validating blob hex strings
var BlobPattern = regexp.MustCompile("^[0-9a-fA-F]+$")

// coerceBlob converts the given value into a valid hex string.
// If the value cannot be converted nil is returned.
func coerceBlob(value any) any {
	switch value := value.(type) {
	case []byte:
		return hex.EncodeToString(value)

	case *[]byte:
		return coerceBlob(*value)

	case string:
		if !BlobPattern.MatchString(value) {
			return nil
		}
		return value

	case *string:
		return coerceBlob(*value)

	default:
		return nil
	}
}

var Blob = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Blob",
	Description: "The `Blob` scalar type represents a binary large object.",
	// Serialize converts the value to a hex string
	Serialize: coerceBlob,
	// ParseValue converts the value to a hex string
	ParseValue: coerceBlob,
	// ParseLiteral converts the ast value to a hex string
	ParseLiteral: func(valueAST ast.Value, variables map[string]any) any {
		switch valueAST := valueAST.(type) {
		case *ast.StringValue:
			return coerceBlob(valueAST.Value)
		default:
			// return nil if the value cannot be parsed
			return nil
		}
	},
})

func parseJSON(valueAST ast.Value, variables map[string]any) any {
	switch valueAST := valueAST.(type) {
	case *ast.ObjectValue:
		out := make(map[string]any)
		for _, f := range valueAST.Fields {
			out[f.Name.Value] = parseJSON(f.Value, variables)
		}
		return out

	case *ast.ListValue:
		out := make([]any, len(valueAST.Values))
		for i, v := range valueAST.Values {
			out[i] = parseJSON(v, variables)
		}
		return out

	case *ast.BooleanValue:
		return graphql.Boolean.ParseLiteral(valueAST, variables)

	case *ast.FloatValue:
		return graphql.Float.ParseLiteral(valueAST, variables)

	case *ast.IntValue:
		return graphql.Int.ParseLiteral(valueAST, variables)

	case *ast.StringValue:
		return graphql.String.ParseLiteral(valueAST, variables)

	case *ast.EnumValue:
		return valueAST.Value

	case *ast.Variable:
		return variables[valueAST.Name.Value]

	default:
		return nil
	}
}

var JSON = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "JSON",
	Description: "The `JSON` scalar type represents a JSON value.",
	// Serialize converts the value to json value
	Serialize: func(value any) any {
		return value
	},
	// ParseValue converts the value to json value
	ParseValue: func(value any) any {
		return value
	},
	// ParseLiteral converts the ast value to a json value
	ParseLiteral: parseJSON,
})

// parseFilterJSON matches parseJSON, except that a bare, unquoted identifier parses into a
// request.FieldReference instead of a plain string, so that a filter such as
// `_alias: {myCount: {_lt: expectedChunks}}` compares against another field of the document.
//
// The shared JSON scalar cannot do this: it is also used for ordering, for filters on
// JSON-typed fields, and for JSON values in mutation inputs, all of which expect a bare
// identifier to stay a plain string.
func parseFilterJSON(valueAST ast.Value, variables map[string]any) any {
	switch valueAST := valueAST.(type) {
	case *ast.ObjectValue:
		out := make(map[string]any)
		for _, f := range valueAST.Fields {
			out[f.Name.Value] = parseFilterJSON(f.Value, variables)
		}
		return out

	case *ast.ListValue:
		out := make([]any, len(valueAST.Values))
		for i, v := range valueAST.Values {
			out[i] = parseFilterJSON(v, variables)
		}
		return out

	case *ast.EnumValue:
		return request.FieldReference{Name: valueAST.Value}

	default:
		return parseJSON(valueAST, variables)
	}
}

// FilterJSON is the type of the `_alias` filter argument.  It is the JSON scalar plus
// support for field-to-field comparison.
var FilterJSON = graphql.NewScalar(graphql.ScalarConfig{
	Name: "FilterJSON",
	Description: "The `FilterJSON` scalar type represents a JSON value in a filter, where a bare " +
		"identifier refers to another field of the same document.",
	Serialize: func(value any) any {
		return value
	},
	ParseValue: func(value any) any {
		return value
	},
	ParseLiteral: parseFilterJSON,
})

// newFilterScalar returns a scalar for use inside filter operator blocks that behaves exactly
// like the given base scalar, except that a bare, unquoted identifier (e.g. `expectedChunks`)
// parses into a request.FieldReference instead of failing type validation.
//
// Two scalars cannot share a name within one schema, so each of these needs its own name.  They
// are only visible in introspection as the argument types of filter operators - the base scalars
// remain in use for field declarations and every other argument.
func newFilterScalar(name string, base *graphql.Scalar) *graphql.Scalar {
	return graphql.NewScalar(graphql.ScalarConfig{
		Name:        name,
		Description: base.Description(),
		Serialize:   base.Serialize,
		ParseValue:  base.ParseValue,
		ParseLiteral: func(valueAST ast.Value, variables map[string]any) any {
			if enum, ok := valueAST.(*ast.EnumValue); ok {
				return request.FieldReference{Name: enum.Value}
			}
			return base.ParseLiteral(valueAST, variables)
		},
	})
}

// The scalars used by the comparison operators (`_eq`, `_neq`, `_gt`, `_geq`, `_lt`, `_leq`)
// of the filter operator blocks in base.go.
var (
	FilterBoolean  = newFilterScalar("FilterBoolean", graphql.Boolean)
	FilterDateTime = newFilterScalar("FilterDateTime", graphql.DateTime)
	FilterFloat32  = newFilterScalar("FilterFloat32", Float32)
	FilterFloat64  = newFilterScalar("FilterFloat64", Float64)
	FilterInt      = newFilterScalar("FilterInt", graphql.Int)
	FilterString   = newFilterScalar("FilterString", graphql.String)
)

// parseAny coerces an AST literal of any kind into a non-nil Go value. It deliberately
// accepts every literal shape (scalars, enums, lists, and objects) so that the `Any`
// scalar passes GraphQL schema validation for any value. The value returned here is only
// used by the validation layer as a non-null gate - the real, field-type-specific coercion
// of a @default value happens in defaultFromAST, which re-parses the raw AST with the
// scalar matching the host field's type.
func parseAny(valueAST ast.Value, variables map[string]any) any {
	switch valueAST := valueAST.(type) {
	case *ast.ObjectValue:
		out := make(map[string]any)
		for _, f := range valueAST.Fields {
			out[f.Name.Value] = parseAny(f.Value, variables)
		}
		return out

	case *ast.ListValue:
		out := make([]any, len(valueAST.Values))
		for i, v := range valueAST.Values {
			out[i] = parseAny(v, variables)
		}
		return out

	case *ast.BooleanValue:
		return valueAST.Value

	case *ast.IntValue:
		return valueAST.Value

	case *ast.FloatValue:
		return valueAST.Value

	case *ast.StringValue:
		return valueAST.Value

	case *ast.EnumValue:
		// Supports bare-word literals such as `UTC_NOW`.
		return valueAST.Value

	case *ast.Variable:
		return variables[valueAST.Name.Value]

	default:
		return nil
	}
}

var Any = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Any",
	Description: "The `Any` scalar type represents any GQL value.",
	// Serialize / ParseValue are passthroughs; `Any` is only ever used as the type of the
	// `@default(value:)` directive argument, which is always supplied as an SDL literal.
	Serialize:    func(value any) any { return value },
	ParseValue:   func(value any) any { return value },
	ParseLiteral: parseAny,
})

func coerceFloat32(value any) any {
	switch value := value.(type) {
	case bool:
		if value {
			return float32(1.0)
		}
		return float32(0.0)
	case *bool:
		if value == nil {
			return nil
		}
		return coerceFloat32(*value)
	case int:
		return float32(value)
	case *int:
		if value == nil {
			return nil
		}
		return coerceFloat32(*value)
	case int8:
		return float32(value)
	case *int8:
		if value == nil {
			return nil
		}
		return coerceFloat32(*value)
	case int16:
		return float32(value)
	case *int16:
		if value == nil {
			return nil
		}
		return coerceFloat32(*value)
	case int32:
		return float32(value)
	case *int32:
		if value == nil {
			return nil
		}
		return coerceFloat32(*value)
	case int64:
		return float32(value)
	case *int64:
		if value == nil {
			return nil
		}
		return coerceFloat32(*value)
	case uint:
		return float32(value)
	case *uint:
		if value == nil {
			return nil
		}
		return coerceFloat32(*value)
	case uint8:
		return float32(value)
	case *uint8:
		if value == nil {
			return nil
		}
		return coerceFloat32(*value)
	case uint16:
		return float32(value)
	case *uint16:
		if value == nil {
			return nil
		}
		return coerceFloat32(*value)
	case uint32:
		return float32(value)
	case *uint32:
		if value == nil {
			return nil
		}
		return coerceFloat32(*value)
	case uint64:
		return float32(value)
	case *uint64:
		if value == nil {
			return nil
		}
		return coerceFloat32(*value)
	case float32:
		return value
	case *float32:
		if value == nil {
			return nil
		}
		return coerceFloat32(*value)
	case float64:
		return float32(value)
	case *float64:
		if value == nil {
			return nil
		}
		return coerceFloat32(*value)
	case string:
		val, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil
		}
		return float32(val)
	case *string:
		if value == nil {
			return nil
		}
		return coerceFloat32(*value)
	}

	// If the value cannot be transformed into an float, return nil instead of '0.0'
	// to denote 'no float found'
	return nil
}

var Float32 = graphql.NewScalar(graphql.ScalarConfig{
	Name: "Float32",
	Description: "The `Float32` scalar type represents signed single-precision fractional " +
		"values as specified by " +
		"[IEEE 754](http://en.wikipedia.org/wiki/IEEE_floating_point). ",
	// Serialize converts the value to float32 value
	Serialize: coerceFloat32,
	// ParseValue converts the value to float32 value
	ParseValue: coerceFloat32,
	// ParseLiteral converts the ast value to a float32 value
	ParseLiteral: func(valueAST ast.Value, variables map[string]any) any {
		switch valueAST := valueAST.(type) {
		case *ast.FloatValue:
			if floatValue, err := strconv.ParseFloat(valueAST.Value, 32); err == nil {
				return float32(floatValue)
			}
		case *ast.IntValue:
			if floatValue, err := strconv.ParseFloat(valueAST.Value, 32); err == nil {
				return float32(floatValue)
			}
		}
		return nil
	},
})

func coerceFloat64(value any) any {
	switch value := value.(type) {
	case bool:
		if value {
			return 1.0
		}
		return 0.0
	case *bool:
		if value == nil {
			return nil
		}
		return coerceFloat64(*value)
	case int:
		return float64(value)
	case *int:
		if value == nil {
			return nil
		}
		return coerceFloat64(*value)
	case int8:
		return float64(value)
	case *int8:
		if value == nil {
			return nil
		}
		return coerceFloat64(*value)
	case int16:
		return float64(value)
	case *int16:
		if value == nil {
			return nil
		}
		return coerceFloat64(*value)
	case int32:
		return float64(value)
	case *int32:
		if value == nil {
			return nil
		}
		return coerceFloat64(*value)
	case int64:
		return float64(value)
	case *int64:
		if value == nil {
			return nil
		}
		return coerceFloat64(*value)
	case uint:
		return float64(value)
	case *uint:
		if value == nil {
			return nil
		}
		return coerceFloat64(*value)
	case uint8:
		return float64(value)
	case *uint8:
		if value == nil {
			return nil
		}
		return coerceFloat64(*value)
	case uint16:
		return float64(value)
	case *uint16:
		if value == nil {
			return nil
		}
		return coerceFloat64(*value)
	case uint32:
		return float64(value)
	case *uint32:
		if value == nil {
			return nil
		}
		return coerceFloat64(*value)
	case uint64:
		return float64(value)
	case *uint64:
		if value == nil {
			return nil
		}
		return coerceFloat64(*value)
	case float32:
		return float64(value)
	case *float32:
		if value == nil {
			return nil
		}
		return coerceFloat64(*value)
	case float64:
		return value
	case *float64:
		if value == nil {
			return nil
		}
		return coerceFloat64(*value)
	case string:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil
		}
		return val
	case *string:
		if value == nil {
			return nil
		}
		return coerceFloat64(*value)
	}

	// If the value cannot be transformed into an float, return nil instead of '0.0'
	// to denote 'no float found'
	return nil
}

var Float64 = graphql.NewScalar(graphql.ScalarConfig{
	Name: "Float64",
	Description: "The `Float64` scalar type represents signed double-precision fractional " +
		"values as specified by " +
		"[IEEE 754](http://en.wikipedia.org/wiki/IEEE_floating_point). ",
	// Serialize converts the value to float64 value
	Serialize: coerceFloat64,
	// ParseValue converts the value to float64 value
	ParseValue: coerceFloat64,
	// ParseLiteral converts the ast value to a float64 value
	ParseLiteral: func(valueAST ast.Value, variables map[string]any) any {
		switch valueAST := valueAST.(type) {
		case *ast.FloatValue:
			if floatValue, err := strconv.ParseFloat(valueAST.Value, 64); err == nil {
				return floatValue
			}
		case *ast.IntValue:
			if floatValue, err := strconv.ParseFloat(valueAST.Value, 64); err == nil {
				return floatValue
			}
		}
		return nil
	},
})
