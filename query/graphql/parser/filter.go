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
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/graphql-go/graphql/language/ast"

	gqlp "github.com/graphql-go/graphql/language/parser"
	gqls "github.com/graphql-go/graphql/language/source"
	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"
)

type EvalContext struct {
	context.Context
}

// Filter contains the parsed condition map to be
// run by the Filter Evaluator.
// @todo: Cache filter structure for faster condition
// evaluation.
type Filter struct {
	// parsed filter conditions
	Conditions map[string]interface{}

	// raw graphql statement
	Statement *ast.ObjectValue
}

// type condition

// NewFilter parses the given GraphQL ObjectValue AST type
// and extracts all the filter conditions into a usable map
func NewFilter(stmt *ast.ObjectValue) (*Filter, error) {
	conditions, err := ParseConditions(stmt)
	if err != nil {
		return nil, err
	}
	return &Filter{
		Statement:  stmt,
		Conditions: conditions,
	}, nil
}

// NewFilterFromString creates a new filter from a string
func NewFilterFromString(body string) (*Filter, error) {
	if !strings.HasPrefix(body, "{") {
		body = "{" + body + "}"
	}
	src := gqls.NewSource(&gqls.Source{Body: []byte(body)})
	p, err := gqlp.MakeParser(src, gqlp.ParseOptions{})
	if err != nil {
		return nil, err
	}
	obj, err := gqlp.ParseObject(p, false)
	if err != nil {
		return nil, err
	}

	return NewFilter(obj)
}

type parseFn func(*ast.ObjectValue) (interface{}, error)

// ParseConditionsInOrder is similar to ParseConditions, except instead
// of returning a map[string]interface{}, we return a []interface{}. This
// is to maintain the ordering info of the statements within the ObjectValue.
// This function is mostly used by the Sort parser, which needs to parse
// conditions in the same way as the Filter object, however the order
// of the arguments is important.
func ParseConditionsInOrder(stmt *ast.ObjectValue) ([]parserTypes.SortCondition, error) {
	cond, err := parseConditionsInOrder(stmt)
	if err != nil {
		return nil, err
	}

	if v, ok := cond.([]parserTypes.SortCondition); ok {
		return v, nil
	}
	return nil, errors.New("Failed to parse statement")
}

func parseConditionsInOrder(stmt *ast.ObjectValue) (interface{}, error) {
	conditions := make([]parserTypes.SortCondition, 0)
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
		case string: // base direction parsed (hopefully, check NameToSortDirection)
			dir, ok := parserTypes.NameToSortDirection[v]
			if !ok {
				return nil, errors.New("Invalid sort direction string")
			}
			conditions = append(conditions, parserTypes.SortCondition{
				Field:     name,
				Direction: dir,
			})

		case []parserTypes.SortCondition: // flatten and incorporate the parsed slice into our current one
			for _, cond := range v {
				// prepend the current field name, to the parsed condition from the slice
				// Eg. sort: {author: {name: ASC, birthday: DESC}}
				// This results in an array of [name, birthday] converted to
				// [author.name, author.birthday].
				// etc.
				cond.Field = fmt.Sprintf("%s.%s", name, cond.Field)
				conditions = append(conditions, cond)
			}

		default:
			return nil, errors.New("Unexpected parsed type for parseConditionInOrder")
		}
	}

	return conditions, nil
}

// func flattenSortDirectionSlice(namespace string, slice []interface{}) []interface{} {
// 	res := make([]interface{}, 0)
// 	for _, v := range slice {
// 		switch n := v.(type) {
// 		case []interface{}:
// 			res = append(res, flatten(namespace, n)...)
// 		default:
// 			res = append(res, n)
// 		}
// 	}
// }

// parseConditions loops over the stmt ObjectValue fields, and extracts
// all the relevant name/value pairs.
func ParseConditions(stmt *ast.ObjectValue) (map[string]interface{}, error) {
	cond, err := parseConditions(stmt)
	if err != nil {
		return nil, err
	}

	if v, ok := cond.(map[string]interface{}); ok {
		return v, nil
	}
	return nil, errors.New("Failed to parse statement")
}

func parseConditions(stmt *ast.ObjectValue) (interface{}, error) {
	conditions := make(map[string]interface{})
	if stmt == nil {
		return conditions, nil
	}
	for _, field := range stmt.Fields {
		name := field.Name.Value
		val, err := parseVal(field.Value, parseConditions)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(name, "_") && name != parserTypes.DocKeyFieldName {
			name = strings.Replace(name, "_", "$", 1)
		}
		conditions[name] = val
	}
	return conditions, nil
}

// parseVal handles all the various input types, and extracts their
// values, with the correct types, into an interface{}.
// recurses on ListValue or ObjectValue
func parseVal(val ast.Value, recurseFn parseFn) (interface{}, error) {
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
	// Ignoring Null input, see:
	// - https://github.com/graphql-go/graphql/issues/178
	// - https://github.com/99designs/gqlgen/issues/1416
	//
	// case "NullValue":
	// 	return nil, nil
	case "ListValue":
		list := make([]interface{}, 0)
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

	return nil, errors.New("Failed to parse condition value from query filter statement")
}

/*
userCollection := db.getCollection("users")
doc := userCollection.NewFromJSON("{
	"title": "Painted House",
	"description": "...",
	"genres": ["bae-123", "bae-def", "bae-456"]
	"author_id": "bae-999",
}")
doc.Save()

doc := document.New(schema).FromJSON

------------------------------------



*/
