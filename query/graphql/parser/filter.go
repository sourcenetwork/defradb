package parser

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/SierraSoftworks/connor"
	"github.com/graphql-go/graphql/language/ast"
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

// parseConditions loops over the stmt ObjectValue fields, and extracts
// all the relevant name/value pairs.
func ParseConditions(stmt *ast.ObjectValue) (map[string]interface{}, error) {
	conditions := make(map[string]interface{})
	for _, field := range stmt.Fields {
		name := field.Name.Value
		val, err := parseVal(field.Value)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(name, "_") {
			name = strings.Replace(name, "_", "$", 1)
		}
		conditions[name] = val
	}
	return conditions, nil
}

// parseVal handles all the various input types, and extracts their
// values, with the correct types, into an interface{}.
// recurses on ListValue or ObjectValue
func parseVal(val ast.Value) (interface{}, error) {
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
			v, err := parseVal(item)
			if err != nil {
				return nil, err
			}
			list = append(list, v)
		}
		return list, nil
	case "ObjectValue":
		conditions, err := ParseConditions(val.(*ast.ObjectValue))
		if err != nil {
			return nil, err
		}
		return conditions, nil
	}

	return nil, errors.New("Failed to parse condition value from query filter statement")
}

// RunFilter runs the given filter expression
// using the document, and evaluates.
func RunFilter(doc map[string]interface{}, filter *Filter, ctx EvalContext) (bool, error) {
	if filter == nil {
		return true, nil
	}
	return connor.Match(filter.Conditions, doc)
}
