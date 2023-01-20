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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	schemaTypes "github.com/sourcenetwork/defradb/query/graphql/schema/types"

	"github.com/sourcenetwork/defradb/client/request"
)

// ParseQuery parses a root ast.Document, and returns a
// formatted Query object.
// Requires a non-nil doc, will error if given a nil doc.
func ParseQuery(schema gql.Schema, doc *ast.Document) (*request.Request, []error) {
	if doc == nil {
		return nil, []error{client.NewErrUninitializeProperty("parseQuery", "doc")}
	}

	r := &request.Request{
		Queries:      make([]*request.OperationDefinition, 0),
		Mutations:    make([]*request.OperationDefinition, 0),
		Subscription: make([]*request.OperationDefinition, 0),
	}

	for _, def := range doc.Definitions {
		astOpDef, isOpDef := def.(*ast.OperationDefinition)
		if !isOpDef {
			continue
		}

		switch astOpDef.Operation {
		case "query":
			parsedQueryOpDef, errs := parseQueryOperationDefinition(schema, astOpDef)
			if errs != nil {
				return nil, errs
			}

			parsedDirectives, err := parseDirectives(astOpDef.Directives)
			if errs != nil {
				return nil, []error{err}
			}
			parsedQueryOpDef.Directives = parsedDirectives

			r.Queries = append(r.Queries, parsedQueryOpDef)

		case "mutation":
			parsedMutationOpDef, err := parseMutationOperationDefinition(schema, astOpDef)
			if err != nil {
				return nil, []error{err}
			}

			parsedDirectives, err := parseDirectives(astOpDef.Directives)
			if err != nil {
				return nil, []error{err}
			}
			parsedMutationOpDef.Directives = parsedDirectives

			r.Mutations = append(r.Mutations, parsedMutationOpDef)

		case "subscription":
			parsedSubscriptionOpDef, err := parseSubscriptionOperationDefinition(schema, astOpDef)
			if err != nil {
				return nil, []error{err}
			}

			parsedDirectives, err := parseDirectives(astOpDef.Directives)
			if err != nil {
				return nil, []error{err}
			}
			parsedSubscriptionOpDef.Directives = parsedDirectives

			r.Subscription = append(r.Subscription, parsedSubscriptionOpDef)

		default:
			return nil, []error{ErrUnknownGQLOperation}
		}
	}

	return r, nil
}

// parseDirectives returns all directives that were found if parsing and validation succeeds,
// otherwise returns the first error that is encountered.
func parseDirectives(astDirectives []*ast.Directive) (request.Directives, error) {
	// Set the default states of the directives if they aren't found and no error(s) occur.
	explainDirective := immutable.None[request.ExplainType]()

	// Iterate through all directives and ensure that the directive we find are validated.
	// - Note: the location we don't need to worry about as the schema takes care of it, as when
	//         request is made there will be a syntax error for directive usage at the wrong location,
	//         unless we add another directive with the same name, for example `@explain` is added
	//         at another location (which we must avoid).
	for _, astDirective := range astDirectives {
		if astDirective == nil {
			return request.Directives{}, errors.New("found a nil directive in the AST")
		}

		if astDirective.Name == nil || astDirective.Name.Value == "" {
			return request.Directives{}, errors.New("found a directive with no name in the AST")
		}

		if astDirective.Name.Value == request.ExplainLabel {
			// Explain directive found, lets parse and validate the directive.
			parsedExplainDirctive, err := parseExplainDirective(astDirective)
			if err != nil {
				return request.Directives{}, err
			}
			explainDirective = parsedExplainDirctive
		}
	}

	return request.Directives{
		ExplainType: explainDirective,
	}, nil
}

// parseExplainDirective parses the explain directive AST and returns an error if the parsing or
// validation goes wrong, otherwise returns the parsed explain type information.
func parseExplainDirective(astDirective *ast.Directive) (immutable.Option[request.ExplainType], error) {
	if len(astDirective.Arguments) == 0 {
		return immutable.Some(request.SimpleExplain), nil
	}

	if len(astDirective.Arguments) != 1 {
		return immutable.None[request.ExplainType](),
			errors.New("invalid number of arguments to an explain request")
	}

	arg := astDirective.Arguments[0]
	if arg.Name.Value != schemaTypes.ExplainArgNameType {
		return immutable.None[request.ExplainType](),
			errors.New("invalid explain request argument")
	}

	switch arg.Value.GetValue() {
	case schemaTypes.ExplainArgSimple:
		return immutable.Some(request.SimpleExplain), nil
	default:
		return immutable.None[request.ExplainType](),
			errors.New("invalid / unknown explain type")
	}
}

// parseQueryOperationDefinition parses the individual GraphQL
// 'query' operations, which there may be multiple of.
func parseQueryOperationDefinition(
	schema gql.Schema,
	def *ast.OperationDefinition,
) (*request.OperationDefinition, []error) {
	qdef := &request.OperationDefinition{
		Selections: make([]request.Selection, len(def.SelectionSet.Selections)),
	}

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
				return nil, ErrFilterMissingArgumentType
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
					return nil, ErrFilterMissingArgumentType
				}
				argTypeObject, ok := argType.(*gql.InputObject)
				if !ok {
					return nil, client.NewErrUnexpectedType[*gql.InputObject]("arg type", argType)
				}
				filterType, ok := getArgumentTypeFromInput(argTypeObject, request.FilterClause)
				if !ok {
					return nil, ErrFilterMissingArgumentType
				}
				filterObjVal, ok := filterArg.Value.(*ast.ObjectValue)
				if !ok {
					return nil, client.NewErrUnexpectedType[*gql.InputObject]("filter arg", filterArg.Value)
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
		return nil, client.NewErrUnhandledType("field", field)
	}
	return fieldObject, nil
}
