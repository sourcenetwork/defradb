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
	"strconv"
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

type parseFn func(*ast.ObjectValue) (any, error)

// ParseConditionsInOrder is similar to ParseConditions, except instead
// of returning a map[string]any, we return a []any. This
// is to maintain the ordering info of the statements within the ObjectValue.
// This function is mostly used by the Order parser, which needs to parse
// conditions in the same way as the Filter object, however the order
// of the arguments is important.
func ParseConditionsInOrder(stmt *ast.ObjectValue) ([]request.OrderCondition, error) {
	cond, err := parseConditionsInOrder(stmt)
	if err != nil {
		return nil, err
	}

	if v, ok := cond.([]request.OrderCondition); ok {
		return v, nil
	}
	return nil, client.NewErrUnexpectedType[[]request.OrderCondition]("condition", cond)
}

func parseConditionsInOrder(stmt *ast.ObjectValue) (any, error) {
	conditions := make([]request.OrderCondition, 0)
	if stmt == nil {
		return conditions, nil
	}
	for _, field := range stmt.Fields {
		name := field.Name.Value
		val, err := parseVal(field.Value, parseConditionsInOrder)
		if err != nil {
			return nil, err
		}

		switch v := val.(type) {
		case string: // base direction parsed (hopefully, check NameToOrderDirection)
			dir, ok := request.NameToOrderDirection[v]
			if !ok {
				return nil, ErrInvalidOrderDirection
			}
			conditions = append(conditions, request.OrderCondition{
				Fields:    []string{name},
				Direction: dir,
			})

		case []request.OrderCondition: // flatten and incorporate the parsed slice into our current one
			for _, cond := range v {
				// prepend the current field name, to the parsed condition from the slice
				// Eg. order: {author: {name: ASC, birthday: DESC}}
				// This results in an array of [name, birthday] converted to
				// [author.name, author.birthday].
				// etc.
				cond.Fields = append([]string{name}, cond.Fields...)
				conditions = append(conditions, cond)
			}

		default:
			return nil, client.NewErrUnhandledType("parseConditionInOrder", val)
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

// parseVal handles all the various input types, and extracts their
// values, with the correct types, into an any.
// recurses on ListValue or ObjectValue
func parseVal(val ast.Value, recurseFn parseFn) (any, error) {
	switch val.GetKind() {
	case "IntValue":
		return strconv.ParseInt(val.GetValue().(string), 10, 64)
	case "FloatValue":
		return strconv.ParseFloat(val.GetValue().(string), 64)
	case "StringValue":
		return val.GetValue().(string), nil
	case "EnumValue":
		return val.GetValue().(string), nil
	case "BooleanValue":
		return val.GetValue().(bool), nil

	case "NullValue":
		return nil, nil

	case "ListValue":
		list := make([]any, 0)
		for _, item := range val.GetValue().([]ast.Value) {
			v, err := parseVal(item, recurseFn)
			if err != nil {
				return nil, err
			}
			list = append(list, v)
		}
		return list, nil
	case "ObjectValue":
		// check recurseFn, its either ParseConditions, or ParseConditionsInOrder
		conditions, err := recurseFn(val.(*ast.ObjectValue))
		if err != nil {
			return nil, err
		}
		return conditions, nil
	}

	return nil, ErrFailedToParseConditionValue
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
			if !found || f.IsObject() {
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
