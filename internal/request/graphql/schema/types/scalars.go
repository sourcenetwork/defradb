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

	"github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"
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

func BlobScalarType() *graphql.Scalar {
	return graphql.NewScalar(graphql.ScalarConfig{
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
}

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

func JSONScalarType() *graphql.Scalar {
	return graphql.NewScalar(graphql.ScalarConfig{
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
}
