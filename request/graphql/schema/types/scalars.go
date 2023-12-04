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

// blobPattern is a regex for validating blob hex strings
var blobPattern = regexp.MustCompile("[0-9a-fA-F]+")

// coerceBlob converts the given value into a valid hex string.
// If the value cannot be converted nil is returned.
func coerceBlob(value any) any {
	switch value := value.(type) {
	case []byte:
		return hex.EncodeToString(value)

	case *[]byte:
		return coerceBlob(*value)

	case string:
		if !blobPattern.MatchString(value) {
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
	// Serialize converts the value to the serialized hex representation
	Serialize: coerceBlob,
	// ParseValue converts the serialized value to the []byte representation
	ParseValue: coerceBlob,
	// ParseLiteral converts the ast value to the []byte representation
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
