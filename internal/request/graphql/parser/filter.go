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
	"strings"

	gql "github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"
	gqlp "github.com/sourcenetwork/graphql-go/language/parser"
	gqls "github.com/sourcenetwork/graphql-go/language/source"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

// type condition

// NewFilter parses the given GraphQL ObjectValue AST type
// and extracts all the filter conditions into a usable map.
func NewFilter(stmt *ast.ObjectValue, inputType gql.Input) (immutable.Option[request.Filter], error) {
	conditions, err := ParseConditions(stmt, inputType)
	if err != nil {
		return immutable.None[request.Filter](), err
	}
	return immutable.Some(request.Filter{
		Conditions: conditions,
	}), nil
}

// NewFilterFromString creates a new filter from a string.
func NewFilterFromString(
	schema gql.Schema,
	collectionType string,
	body string,
) (immutable.Option[request.Filter], error) {
	if !strings.HasPrefix(body, "{") {
		body = "{" + body + "}"
	}
	src := gqls.NewSource(&gqls.Source{Body: []byte(body)})
	p, err := gqlp.MakeParser(src, gqlp.ParseOptions{})
	if err != nil {
		return immutable.None[request.Filter](), err
	}
	obj, err := gqlp.ParseObject(p, false)
	if err != nil {
		return immutable.None[request.Filter](), err
	}

	parentFieldType := gql.GetFieldDef(schema, schema.QueryType(), collectionType)
	filterType, ok := getArgumentType(parentFieldType, request.FilterClause)
	if !ok {
		return immutable.None[request.Filter](), ErrFilterMissingArgumentType
	}
	return NewFilter(obj, filterType)
}

// parseConditions loops over the stmt ObjectValue fields, and extracts
// all the relevant name/value pairs.
func ParseConditions(stmt *ast.ObjectValue, inputType gql.Input) (map[string]any, error) {
	cond := gql.ValueFromAST(stmt, inputType, nil)
	if cond == nil {
		return nil, ErrFailedToParseConditionsFromAST
	}
	if v, ok := cond.(map[string]any); ok {
		return v, nil
	}
	return nil, client.NewErrUnexpectedType[map[string]any]("condition", cond)
}

// ParseFilterFieldsForDescription parses the fields that are defined in the SchemaDescription
// from the filter conditions“
func ParseFilterFieldsForDescription(
	conditions map[string]any,
	col client.CollectionDefinition,
) ([]client.FieldDefinition, error) {
	return parseFilterFieldsForDescriptionMap(conditions, col)
}

func parseFilterFieldsForDescriptionMap(
	conditions map[string]any,
	col client.CollectionDefinition,
) ([]client.FieldDefinition, error) {
	fields := make([]client.FieldDefinition, 0)
	for k, v := range conditions {
		switch k {
		case "_or", "_and":
			conds := v.([]any)
			parsedFileds, err := parseFilterFieldsForDescriptionSlice(conds, col)
			if err != nil {
				return nil, err
			}
			fields = append(fields, parsedFileds...)
		case "_not":
			conds := v.(map[string]any)
			parsedFileds, err := parseFilterFieldsForDescriptionMap(conds, col)
			if err != nil {
				return nil, err
			}
			fields = append(fields, parsedFileds...)
		default:
			f, found := col.GetFieldByName(k)
			if !found || f.Kind.IsObject() {
				continue
			}
			fields = append(fields, f)
		}
	}
	return fields, nil
}

func parseFilterFieldsForDescriptionSlice(
	conditions []any,
	schema client.CollectionDefinition,
) ([]client.FieldDefinition, error) {
	fields := make([]client.FieldDefinition, 0)
	for _, v := range conditions {
		switch cond := v.(type) {
		case map[string]any:
			parsedFields, err := parseFilterFieldsForDescriptionMap(cond, schema)
			if err != nil {
				return nil, err
			}
			fields = append(fields, parsedFields...)
		default:
			return nil, ErrInvalidFilterConditions
		}
	}
	return fields, nil
}
