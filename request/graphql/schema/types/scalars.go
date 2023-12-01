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

	"github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"
)

var BlobScalarType = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Blob",
	Description: "The `Blob` scalar type represents a binary large object.",
	Serialize: func(value any) any {
		switch value := value.(type) {
		case []byte:
			return hex.EncodeToString(value)
		case *[]byte:
			return hex.EncodeToString(*value)
		default:
			// return nil if the value cannot be serialized
			return nil
		}
	},
	ParseValue: func(value any) any {
		switch value := value.(type) {
		case string:
			data, err := hex.DecodeString(value)
			if err != nil {
				return nil
			}
			return data
		case *string:
			data, err := hex.DecodeString(*value)
			if err != nil {
				// the error cannot be handled due to
				// the design of graphql-go scalars
				//
				// return nil if the value cannot be parsed
				return nil
			}
			return data
		default:
			return nil
		}
	},
	ParseLiteral: func(valueAST ast.Value) any {
		switch valueAST := valueAST.(type) {
		case *ast.StringValue:
			data, err := hex.DecodeString(valueAST.Value)
			if err != nil {
				// the error cannot be handled due to
				// the design of graphql-go scalars
				//
				// return nil if the value cannot be parsed
				return nil
			}
			return data
		default:
			// return nil if the value cannot be parsed
			return nil
		}
	},
})
