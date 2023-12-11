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
	"math/big"
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

var BlobScalarType = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Blob",
	Description: "The `Blob` scalar type represents a binary large object.",
	// Serialize converts the value to a hex string
	Serialize: coerceBlob,
	// ParseValue converts the value to a hex string
	ParseValue: coerceBlob,
	// ParseLiteral converts the ast value to a hex string
	ParseLiteral: func(valueAST ast.Value) any {
		switch valueAST := valueAST.(type) {
		case *ast.StringValue:
			return coerceBlob(valueAST.Value)
		default:
			// return nil if the value cannot be parsed
			return nil
		}
	},
})

var BigIntPattern = regexp.MustCompile("^[0-9]+$")

var BigIntScalarType = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "BigInt",
	Description: "The `BigInt` scalar type represents an arbitrary precision integer.",
	// Serialize converts the value to a string
	Serialize: coerceBigInt,
	// ParseValue converts the value to a string
	ParseValue: coerceBigInt,
	// ParseLiteral converts the ast value to a string
	ParseLiteral: func(valueAST ast.Value) any {
		switch valueAST := valueAST.(type) {
		case *ast.IntValue:
			return coerceBigInt(valueAST.Value)
		case *ast.FloatValue:
			return coerceBigInt(valueAST.Value)
		case *ast.StringValue:
			return coerceBigInt(valueAST.Value)
		default:
			// return nil if the value cannot be parsed
			return nil
		}
	},
})

// coerceBigInt converts the given value into a valid BigInt.
// If the value cannot be converted nil is returned.
func coerceBigInt(value any) any {
	switch value := value.(type) {
	case float32:
		return big.NewInt(int64(value)).String()

	case *float32:
		return coerceBigInt(*value)

	case float64:
		return big.NewInt(int64(value)).String()

	case *float64:
		return coerceBigInt(*value)

	case int:
		return big.NewInt(int64(value)).String()

	case *int:
		return coerceBigInt(*value)

	case int16:
		return big.NewInt(int64(value)).String()

	case *int16:
		return coerceBigInt(*value)

	case int32:
		return big.NewInt(int64(value)).String()

	case *int32:
		return coerceBigInt(*value)

	case int64:
		return big.NewInt(int64(value)).String()

	case *int64:
		return coerceBigInt(*value)

	case uint:
		return big.NewInt(int64(value)).String()

	case *uint:
		return coerceBigInt(*value)

	case uint16:
		return big.NewInt(int64(value)).String()

	case *uint16:
		return coerceBigInt(*value)

	case uint32:
		return big.NewInt(int64(value)).String()

	case *uint32:
		return coerceBigInt(*value)

	case uint64:
		return big.NewInt(int64(value)).String()

	case *uint64:
		return coerceBigInt(*value)

	case string:
		if !BigIntPattern.MatchString(value) {
			return nil
		}
		return value

	case *string:
		return coerceBlob(*value)

	default:
		return nil
	}
}
