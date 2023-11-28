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

var BytesScalarType = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Bytes",
	Description: "The `Bytes` scalar type represents an array of bytes.",
	Serialize: func(value any) any {
		switch value := value.(type) {
		case []byte:
			return hex.EncodeToString(value)
		default:
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
				return nil
			}
			return data
		default:
			return nil
		}
	},
})
