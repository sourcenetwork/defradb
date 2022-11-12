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
	"github.com/graphql-go/graphql/language/ast"

	"github.com/sourcenetwork/defradb/client/request"
)

// parseSubscriptionOperationDefinition parses the individual GraphQL
// 'subcription' operations, which there may be multiple of.
func parseSubscriptionOperationDefinition(def *ast.OperationDefinition) (*request.OperationDefinition, error) {
	sdef := &request.OperationDefinition{
		Selections: make([]request.Selection, len(def.SelectionSet.Selections)),
	}

	sdef.IsExplain = parseExplainDirective(def.Directives)

	for i, selection := range def.SelectionSet.Selections {
		switch node := selection.(type) {
		case *ast.Field:
			sub, err := parseSubscription(node)
			if err != nil {
				return nil, err
			}

			sdef.Selections[i] = sub
		}
	}
	return sdef, nil
}

// parseSubscription parses a typed subscription field
// which includes sub fields, and may include
// filters, IDs, etc.
func parseSubscription(field *ast.Field) (*request.ObjectSubscription, error) {
	sub := &request.ObjectSubscription{
		Field: request.Field{
			Name:  field.Name.Value,
			Alias: getFieldAlias(field),
		},
	}

	sub.Collection = sub.Name

	// parse arguments
	for _, argument := range field.Arguments {
		prop := argument.Name.Value
		if prop == request.FilterClause { // parse filter
			obj := argument.Value.(*ast.ObjectValue)
			filter, err := NewFilter(obj)
			if err != nil {
				return nil, err
			}

			sub.Filter = filter
		}
	}

	var err error
	sub.Fields, err = parseSelectFields(request.ObjectSelection, field.SelectionSet)
	return sub, err
}
