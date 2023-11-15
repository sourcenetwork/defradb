// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package parser

import (
	gql "github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"
)

// IsIntrospectionQuery parses a root ast.Document and determines if it is an
// introspection query. This is used to determine if the query should be
// executed against the schema or the database.
func IsIntrospectionQuery(schema gql.Schema, doc *ast.Document) bool {
	for _, def := range doc.Definitions {
		astOpDef, isOpDef := def.(*ast.OperationDefinition)
		if !isOpDef {
			continue
		}

		if astOpDef.Operation == ast.OperationTypeQuery {
			for _, selection := range astOpDef.SelectionSet.Selections {
				switch node := selection.(type) {
				case *ast.Field:
					if node.Name.Value == "__schema" || node.Name.Value == "__type" {
						return true
					}
				default:
					return false
				}
			}
		}
	}

	return false
}
