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

	"github.com/sourcenetwork/defradb/client/request"
)

// parseSubscriptionOperationDefinition parses the individual GraphQL
// 'subcription' operations, which there may be multiple of.
func parseSubscriptionOperationDefinition(
	schema gql.Schema,
	def *ast.OperationDefinition,
) (*request.OperationDefinition, error) {
	sdef := &request.OperationDefinition{
		Selections: make([]request.Selection, len(def.SelectionSet.Selections)),
	}

	for i, selection := range def.SelectionSet.Selections {
		switch node := selection.(type) {
		case *ast.Field:
			sub, err := parseSubscription(schema, node)
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
func parseSubscription(schema gql.Schema, field *ast.Field) (*request.ObjectSubscription, error) {
	sub := &request.ObjectSubscription{
		Field: request.Field{
			Name:  field.Name.Value,
			Alias: getFieldAlias(field),
		},
	}

	sub.Collection = sub.Name

	fieldDef := gql.GetFieldDef(schema, schema.QueryType(), field.Name.Value)

	for _, argument := range field.Arguments {
		prop := argument.Name.Value
		if prop == request.FilterClause {
			filterType, ok := getArgumentType(fieldDef, request.FilterClause)
			if !ok {
				return nil, ErrFilterMissingArgumentType
			}
			obj := argument.Value.(*ast.ObjectValue)
			filter, err := NewFilter(obj, filterType)
			if err != nil {
				return nil, err
			}

			sub.Filter = filter
		}
	}

	// parse field selections
	fieldObject, err := typeFromFieldDef(fieldDef)
	if err != nil {
		return nil, err
	}

	sub.Fields, err = parseSelectFields(schema, request.ObjectSelection, fieldObject, field.SelectionSet)
	return sub, err
}
