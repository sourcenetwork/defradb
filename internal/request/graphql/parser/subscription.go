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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
)

// parseSubscriptionOperationDefinition parses the individual GraphQL
// 'subcription' operations, which there may be multiple of.
func parseSubscriptionOperationDefinition(
	exe *gql.ExecutionContext,
	collectedFields map[string][]*ast.Field,
) (*request.OperationDefinition, error) {
	var selections []request.Selection
	for _, fields := range collectedFields {
		for _, node := range fields {
			sub, err := parseSubscription(exe, node)
			if err != nil {
				return nil, err
			}
			selections = append(selections, sub)
		}
	}
	return &request.OperationDefinition{
		Selections: selections,
	}, nil
}

// parseSubscription parses a typed subscription field
// which includes sub fields, and may include
// filters, IDs, etc.
func parseSubscription(exe *gql.ExecutionContext, field *ast.Field) (*request.ObjectSubscription, error) {
	sub := &request.ObjectSubscription{
		Field: request.Field{
			Name:  field.Name.Value,
			Alias: getFieldAlias(field),
		},
	}

	sub.Collection = sub.Name

	fieldDef := gql.GetFieldDef(exe.Schema, exe.Schema.QueryType(), field.Name.Value)
	arguments := gql.GetArgumentValues(fieldDef.Args, field.Arguments, exe.VariableValues)

	if v, ok := arguments[request.FilterClause].(map[string]any); ok {
		sub.Filter = immutable.Some(request.Filter{Conditions: v})
	}

	// parse field selections
	fieldObject, err := typeFromFieldDef(fieldDef)
	if err != nil {
		return nil, err
	}

	sub.Fields, err = parseSelectFields(exe, fieldObject, field.SelectionSet)
	if err != nil {
		return nil, err
	}
	return sub, err
}
