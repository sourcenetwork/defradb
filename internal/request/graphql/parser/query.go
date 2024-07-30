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

	gql "github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"

	"github.com/sourcenetwork/defradb/client/request"
)

// parseQueryOperationDefinition parses the individual GraphQL
// 'query' operations, which there may be multiple of.
func parseQueryOperationDefinition(
	schema gql.Schema,
	def *ast.OperationDefinition,
) (*request.OperationDefinition, []error) {
	qdef := &request.OperationDefinition{
		Name:       def.Name.Value,
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
					ChildSelect: request.ChildSelect{
						Fields: []request.Selection{
							parsed,
						},
					},
				}
			} else {
				// the query doesn't match a reserve name
				// so its probably a generated query
				parsed, err := parseSelect(schema, schema.QueryType(), node, i)
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
	parent *gql.Object,
	field *ast.Field,
	index int,
) (*request.Select, error) {
	slct := &request.Select{
		Field: request.Field{
			Name:  field.Name.Value,
			Alias: getFieldAlias(field),
		},
	}

	fieldDef := gql.GetFieldDef(schema, parent, slct.Name)

	// parse arguments
	for _, argument := range field.Arguments {
		prop := argument.Name.Value
		astValue := argument.Value

		// parse filter
		switch prop {
		case request.FilterClause:
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
		case request.DocIDArgName: // parse single DocID field
			docIDValue := astValue.(*ast.StringValue)
			slct.DocIDs = immutable.Some([]string{docIDValue.Value})
		case request.DocIDsArgName:
			docIDValues := astValue.(*ast.ListValue).Values
			docIDs := make([]string, len(docIDValues))
			for i, value := range docIDValues {
				docIDs[i] = value.(*ast.StringValue).Value
			}
			slct.DocIDs = immutable.Some(docIDs)
		case request.Cid: // parse single CID query field
			val := astValue.(*ast.StringValue)
			slct.CID = immutable.Some(val.Value)
		case request.LimitClause: // parse limit/offset
			val := astValue.(*ast.IntValue)
			limit, err := strconv.ParseUint(val.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			slct.Limit = immutable.Some(limit)
		case request.OffsetClause: // parse limit/offset
			val := astValue.(*ast.IntValue)
			offset, err := strconv.ParseUint(val.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			slct.Offset = immutable.Some(offset)
		case request.OrderClause: // parse order by
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
		case request.GroupByClause:
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
		case request.ShowDeleted:
			val := astValue.(*ast.BooleanValue)
			slct.ShowDeleted = val.Value
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

	slct.Fields, err = parseSelectFields(schema, fieldObject, field.SelectionSet)
	if err != nil {
		return nil, err
	}

	return slct, err
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
				Filterable: request.Filterable{
					Filter: filter,
				},
				Limitable: request.Limitable{
					Limit: limit,
				},
				Offsetable: request.Offsetable{
					Offset: offset,
				},
				Orderable: request.Orderable{
					OrderBy: order,
				},
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
