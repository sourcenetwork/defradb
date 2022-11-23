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

	gql "github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
)

// ParseQuery parses a root ast.Document, and returns a
// formatted Query object.
// Requires a non-nil doc, will error if given a nil doc.
func ParseQuery(schema gql.Schema, doc *ast.Document) (*request.Request, []error) {
	if doc == nil {
		return nil, []error{errors.New("parseQuery requires a non-nil ast.Document")}
	}
	r := &request.Request{
		Queries:      make([]*request.OperationDefinition, 0),
		Mutations:    make([]*request.OperationDefinition, 0),
		Subscription: make([]*request.OperationDefinition, 0),
	}

	for _, def := range doc.Definitions {
		switch node := def.(type) {
		case *ast.OperationDefinition:
			switch node.Operation {
			case "query":
				// parse query operation definition.
				qdef, err := parseQueryOperationDefinition(schema, node)
				if err != nil {
					return nil, err
				}
				r.Queries = append(r.Queries, qdef)
			case "mutation":
				// parse mutation operation definition.
				mdef, err := parseMutationOperationDefinition(schema, node)
				if err != nil {
					return nil, []error{err}
				}
				r.Mutations = append(r.Mutations, mdef)
			case "subscription":
				// parse subscription operation definition.
				sdef, err := parseSubscriptionOperationDefinition(schema, node)
				if err != nil {
					return nil, []error{err}
				}
				r.Subscription = append(r.Subscription, sdef)
			default:
				return nil, []error{errors.New("unknown GraphQL operation type")}
			}
		}
	}

	return r, nil
}

// parseExplainDirective returns true if we parsed / detected the explain directive label
// in this ast, and false otherwise.
func parseExplainDirective(directives []*ast.Directive) bool {
	// Iterate through all directives and ensure that the directive is at there.
	// - Note: the location we don't need to worry about as the schema takes care of it, as when
	//         request is made there will be a syntax error for directive usage at the wrong location,
	//         unless we add another directive named `@explain` at another location (which we should not).
	for _, directive := range directives {
		// The arguments pased to the directive are at `directive.Arguments`.
		if directive.Name.Value == request.ExplainLabel {
			return true
		}
	}

	return false
}

// parseQueryOperationDefinition parses the individual GraphQL
// 'query' operations, which there may be multiple of.
func parseQueryOperationDefinition(
	schema gql.Schema,
	def *ast.OperationDefinition) (*request.OperationDefinition, []error) {
	qdef := &request.OperationDefinition{
		Selections: make([]request.Selection, len(def.SelectionSet.Selections)),
	}

	qdef.IsExplain = parseExplainDirective(def.Directives)

	for i, selection := range def.SelectionSet.Selections {
		var parsedSelection request.Selection
		switch node := selection.(type) {
		case *ast.Field:
			if _, isCommitQuery := request.CommitQueries[node.Name.Value]; isCommitQuery {
				parsed, err := parseCommitSelect(schema, schema.QueryType(), node)
				if err != nil {
					return nil, []error{err}
				}

				parsedSelection = parsed
			} else if _, isAggregate := request.Aggregates[node.Name.Value]; isAggregate {
				parsed, err := parseAggregate(schema, schema.QueryType(), node, i)
				if err != nil {
					return nil, []error{err}
				}

				// Top-level aggregates must be wrapped in a top-level Select for now
				parsedSelection = &request.Select{
					Field: request.Field{
						Name:  parsed.Name,
						Alias: parsed.Alias,
					},
					Fields: []request.Selection{
						parsed,
					},
				}
			} else {
				// the query doesn't match a reserve name
				// so its probably a generated query
				parsed, err := parseSelect(schema, request.ObjectSelection, schema.QueryType(), node, i)
				if err != nil {
					return nil, []error{err}
				}

				errors := parsed.Validate()
				if len(errors) > 0 {
					return nil, errors
				}

				parsedSelection = parsed
			}

			qdef.Selections[i] = parsedSelection
		}
	}
	return qdef, nil
}

// @todo: Create separate select parse functions
// for generated object queries, and general
// API queries

// parseSelect parses a typed selection field
// which includes sub fields, and may include
// filters, limits, orders, etc..
func parseSelect(
	schema gql.Schema,
	rootType request.SelectionType,
	parent *gql.Object,
	field *ast.Field,
	index int,
) (*request.Select, error) {
	slct := &request.Select{
		Field: request.Field{
			Name:  field.Name.Value,
			Alias: getFieldAlias(field),
		},
		Root: rootType,
	}

	fieldDef := gql.GetFieldDef(schema, parent, slct.Name)

	// parse arguments
	for _, argument := range field.Arguments {
		prop := argument.Name.Value
		astValue := argument.Value

		// parse filter
		if prop == request.FilterClause {
			obj := astValue.(*ast.ObjectValue)
			filterType, ok := getArgumentType(fieldDef, request.FilterClause)
			if !ok {
				return nil, errors.New("couldn't get argument type for filter")
			}
			filter, err := NewFilter(obj, filterType)
			if err != nil {
				return slct, err
			}

			slct.Filter = filter
		} else if prop == request.DocKey { // parse single dockey query field
			val := astValue.(*ast.StringValue)
			slct.DocKeys = immutable.Some([]string{val.Value})
		} else if prop == request.DocKeys {
			docKeyValues := astValue.(*ast.ListValue).Values
			docKeys := make([]string, len(docKeyValues))
			for i, value := range docKeyValues {
				docKeys[i] = value.(*ast.StringValue).Value
			}
			slct.DocKeys = immutable.Some(docKeys)
		} else if prop == request.Cid { // parse single CID query field
			val := astValue.(*ast.StringValue)
			slct.CID = immutable.Some(val.Value)
		} else if prop == request.LimitClause { // parse limit/offset
			val := astValue.(*ast.IntValue)
			limit, err := strconv.ParseUint(val.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			slct.Limit = immutable.Some(limit)
		} else if prop == request.OffsetClause { // parse limit/offset
			val := astValue.(*ast.IntValue)
			offset, err := strconv.ParseUint(val.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			slct.Offset = immutable.Some(offset)
		} else if prop == request.OrderClause { // parse order by
			obj := astValue.(*ast.ObjectValue)
			cond, err := ParseConditionsInOrder(obj)
			if err != nil {
				return nil, err
			}
			slct.OrderBy = immutable.Some(
				request.OrderBy{
					Conditions: cond,
				},
			)
		} else if prop == request.GroupByClause {
			obj := astValue.(*ast.ListValue)
			fields := make([]string, 0)
			for _, v := range obj.Values {
				fields = append(fields, v.GetValue().(string))
			}

			slct.GroupBy = immutable.Some(
				request.GroupBy{
					Fields: fields,
				},
			)
		}
	}

	// if theres no field selections, just return
	if field.SelectionSet == nil {
		return slct, nil
	}

	// parse field selections
	fieldObject, err := typeFromFieldDef(fieldDef)
	if err != nil {
		return nil, err
	}

	slct.Fields, err = parseSelectFields(schema, slct.Root, fieldObject, field.SelectionSet)
	if err != nil {
		return nil, err
	}

	return slct, err
}

func getFieldAlias(field *ast.Field) immutable.Option[string] {
	if field.Alias == nil {
		return immutable.None[string]()
	}
	return immutable.Some(field.Alias.Value)
}

func parseSelectFields(
	schema gql.Schema,
	root request.SelectionType,
	parent *gql.Object,
	fields *ast.SelectionSet) ([]request.Selection, error) {
	selections := make([]request.Selection, len(fields.Selections))
	// parse field selections
	for i, selection := range fields.Selections {
		switch node := selection.(type) {
		case *ast.Field:
			if _, isAggregate := request.Aggregates[node.Name.Value]; isAggregate {
				s, err := parseAggregate(schema, parent, node, i)
				if err != nil {
					return nil, err
				}
				selections[i] = s
			} else if node.SelectionSet == nil { // regular field
				selections[i] = parseField(node)
			} else { // sub type with extra fields
				subroot := root
				switch node.Name.Value {
				case request.VersionFieldName:
					subroot = request.CommitSelection
				}

				s, err := parseSelect(schema, subroot, parent, node, i)
				if err != nil {
					return nil, err
				}
				selections[i] = s
			}
		}
	}

	return selections, nil
}

// parseField simply parses the Name/Alias
// into a Field type
func parseField(field *ast.Field) *request.Field {
	return &request.Field{
		Name:  field.Name.Value,
		Alias: getFieldAlias(field),
	}
}

func parseAggregate(schema gql.Schema, parent *gql.Object, field *ast.Field, index int) (*request.Aggregate, error) {
	targets := make([]*request.AggregateTarget, len(field.Arguments))

	for i, argument := range field.Arguments {
		switch argumentValue := argument.Value.GetValue().(type) {
		case string:
			targets[i] = &request.AggregateTarget{
				HostName: argumentValue,
			}
		case []*ast.ObjectField:
			hostName := argument.Name.Value
			var childName string
			var filter immutable.Option[request.Filter]
			var limit immutable.Option[uint64]
			var offset immutable.Option[uint64]
			var order immutable.Option[request.OrderBy]

			fieldArg, hasFieldArg := tryGet(argumentValue, request.FieldName)
			if hasFieldArg {
				if innerPathStringValue, isString := fieldArg.Value.GetValue().(string); isString {
					childName = innerPathStringValue
				}
			}

			filterArg, hasFilterArg := tryGet(argumentValue, request.FilterClause)
			if hasFilterArg {
				fieldDef := gql.GetFieldDef(schema, parent, field.Name.Value)
				argType, ok := getArgumentType(fieldDef, hostName)
				if !ok {
					return nil, errors.New("couldn't get argument type for filter")
				}
				argTypeObject, ok := argType.(*gql.InputObject)
				if !ok {
					return nil, errors.New("expected arg type to be object")
				}
				filterType, ok := getArgumentTypeFromInput(argTypeObject, request.FilterClause)
				if !ok {
					return nil, errors.New("couldn't get argument type for filter")
				}
				filterObjVal, ok := filterArg.Value.(*ast.ObjectValue)
				if !ok {
					return nil, errors.New("couldn't get object value type for filter")
				}
				filterValue, err := NewFilter(filterObjVal, filterType)
				if err != nil {
					return nil, err
				}
				filter = filterValue
			}

			limitArg, hasLimitArg := tryGet(argumentValue, request.LimitClause)
			if hasLimitArg {
				limitValue, err := strconv.ParseUint(limitArg.Value.(*ast.IntValue).Value, 10, 64)
				if err != nil {
					return nil, err
				}
				limit = immutable.Some(limitValue)
			}

			offsetArg, hasOffsetArg := tryGet(argumentValue, request.OffsetClause)
			if hasOffsetArg {
				offsetValue, err := strconv.ParseUint(offsetArg.Value.(*ast.IntValue).Value, 10, 64)
				if err != nil {
					return nil, err
				}
				offset = immutable.Some(offsetValue)
			}

			orderArg, hasOrderArg := tryGet(argumentValue, request.OrderClause)
			if hasOrderArg {
				switch orderArgValue := orderArg.Value.(type) {
				case *ast.EnumValue:
					// For inline arrays the order arg will be a simple enum declaring the order direction
					orderDirectionString := orderArgValue.Value
					orderDirection := request.OrderDirection(orderDirectionString)

					order = immutable.Some(
						request.OrderBy{
							Conditions: []request.OrderCondition{
								{
									Direction: orderDirection,
								},
							},
						},
					)

				case *ast.ObjectValue:
					// For relations the order arg will be the complex order object as used by the host object
					// for non-aggregate ordering

					// We use the parser package parsing for convienience here
					orderConditions, err := ParseConditionsInOrder(orderArgValue)
					if err != nil {
						return nil, err
					}

					order = immutable.Some(
						request.OrderBy{
							Conditions: orderConditions,
						},
					)
				}
			}

			targets[i] = &request.AggregateTarget{
				HostName:  hostName,
				ChildName: immutable.Some(childName),
				Filter:    filter,
				Limit:     limit,
				Offset:    offset,
				OrderBy:   order,
			}
		}
	}

	return &request.Aggregate{
		Field: request.Field{
			Name:  field.Name.Value,
			Alias: getFieldAlias(field),
		},
		Targets: targets,
	}, nil
}

func tryGet(fields []*ast.ObjectField, name string) (*ast.ObjectField, bool) {
	for _, field := range fields {
		if field.Name.Value == name {
			return field, true
		}
	}
	return nil, false
}

func getArgumentType(field *gql.FieldDefinition, name string) (gql.Input, bool) {
	for _, arg := range field.Args {
		if arg.Name() == name {
			return arg.Type, true
		}
	}
	return nil, false
}

func getArgumentTypeFromInput(input *gql.InputObject, name string) (gql.Input, bool) {
	for fname, ftype := range input.Fields() {
		if fname == name {
			return ftype.Type, true
		}
	}
	return nil, false
}

// typeFromFieldDef will return the output gql.Object type from the given field.
// The return type may be a gql.Object or a gql.List, if it is a List type, we
// need to get the concrete "OfType".
func typeFromFieldDef(field *gql.FieldDefinition) (*gql.Object, error) {
	var fieldObject *gql.Object
	switch ftype := field.Type.(type) {
	case *gql.Object:
		fieldObject = ftype
	case *gql.List:
		fieldObject = ftype.OfType.(*gql.Object)
	default:
		return nil, errors.New("couldn't get field object from definition")
	}
	return fieldObject, nil
}
