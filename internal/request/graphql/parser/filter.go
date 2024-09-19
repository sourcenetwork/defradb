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

// ParseConditionsInOrder is similar to ParseConditions, except instead
// of returning a map[string]any, we return a []any. This
// is to maintain the ordering info of the statements within the ObjectValue.
// This function is mostly used by the Order parser, which needs to parse
// conditions in the same way as the Filter object, however the order
// of the arguments is important.
func ParseConditionsInOrder(stmt *ast.ObjectValue, args map[string]any) ([]request.OrderCondition, error) {
	conditions := make([]request.OrderCondition, 0)
	if stmt == nil {
		return conditions, nil
	}
	for _, field := range stmt.Fields {
		switch v := args[field.Name.Value].(type) {
		case int: // base direction parsed (hopefully, check NameToOrderDirection)
			dir, err := parseOrderDirection(v)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, request.OrderCondition{
				Fields:    []string{field.Name.Value},
				Direction: dir,
			})

		case map[string]any: // flatten and incorporate the parsed slice into our current one
			sub, err := ParseConditionsInOrder(field.Value.(*ast.ObjectValue), v)
			if err != nil {
				return nil, err
			}
			for _, cond := range sub {
				// prepend the current field name, to the parsed condition from the slice
				// Eg. order: {author: {name: ASC, birthday: DESC}}
				// This results in an array of [name, birthday] converted to
				// [author.name, author.birthday].
				// etc.
				cond.Fields = append([]string{field.Name.Value}, cond.Fields...)
				conditions = append(conditions, cond)
			}

		case nil:
			continue // ignore nil filter input

		default:
			return nil, client.NewErrUnhandledType("parseConditionInOrder", v)
		}
	}

	return conditions, nil
}

// parseConditions loops over the stmt ObjectValue fields, and extracts
// all the relevant name/value pairs.
func ParseConditions(stmt *ast.ObjectValue, inputType gql.Input) (map[string]any, error) {
	cond, err := parseConditions(stmt, inputType)
	if err != nil {
		return nil, err
	}

	if v, ok := cond.(map[string]any); ok {
		return v, nil
	}
	return nil, client.NewErrUnexpectedType[map[string]any]("condition", cond)
}

func parseConditions(stmt *ast.ObjectValue, inputArg gql.Input) (any, error) {
	val := gql.ValueFromAST(stmt, inputArg, nil)
	if val == nil {
		return nil, ErrFailedToParseConditionsFromAST
	}
	return val, nil
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

func parseOrderDirection(v int) (request.OrderDirection, error) {
	switch v {
	case 0:
		return request.ASC, nil

	case 1:
		return request.DESC, nil

	default:
		return request.ASC, ErrInvalidOrderDirection
	}
}
