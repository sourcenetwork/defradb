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

// parseQueryOperationDefinition parses the individual GraphQL
// 'query' operations, which there may be multiple of.
func parseQueryOperationDefinition(
	exe *gql.ExecutionContext,
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
				parsed, err := parseCommitSelect(exe, exe.Schema.QueryType(), node)
				if err != nil {
					return nil, []error{err}
				}

				parsedSelection = parsed
			} else if _, isAggregate := request.Aggregates[node.Name.Value]; isAggregate {
				parsed, err := parseAggregate(exe, exe.Schema.QueryType(), node)
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
				parsed, err := parseSelect(exe, exe.Schema.QueryType(), node)
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
	exe *gql.ExecutionContext,
	parent *gql.Object,
	field *ast.Field,
) (*request.Select, error) {
	slct := &request.Select{
		Field: request.Field{
			Name:  field.Name.Value,
			Alias: getFieldAlias(field),
		},
	}

	fieldDef := gql.GetFieldDef(exe.Schema, parent, field.Name.Value)
	arguments := gql.GetArgumentValues(fieldDef.Args, field.Arguments, exe.VariableValues)

	// parse arguments
	for _, argument := range field.Arguments {
		name := argument.Name.Value
		value := arguments[name]

		// parse filter
		switch name {
		case request.FilterClause:
			slct.Filter = immutable.Some(request.Filter{
				Conditions: value.(map[string]any),
			})
		case request.DocIDArgName: // parse single DocID field
			slct.DocIDs = immutable.Some([]string{value.(string)})
		case request.DocIDsArgName:
			docIDValues := value.([]any)
			docIDs := make([]string, len(docIDValues))
			for i, value := range docIDValues {
				docIDs[i] = value.(string)
			}
			slct.DocIDs = immutable.Some(docIDs)
		case request.Cid: // parse single CID query field
			slct.CID = immutable.Some(value.(string))
		case request.LimitClause: // parse limit/offset
			slct.Limit = immutable.Some(uint64(value.(int32)))
		case request.OffsetClause: // parse limit/offset
			slct.Offset = immutable.Some(uint64(value.(int32)))
		case request.OrderClause: // parse order by
			conditionsAST := argument.Value.(*ast.ObjectValue)
			conditionsValue := value.(map[string]any)
			conditions, err := ParseConditionsInOrder(conditionsAST, conditionsValue)
			if err != nil {
				return nil, err
			}
			slct.OrderBy = immutable.Some(request.OrderBy{
				Conditions: conditions,
			})
		case request.GroupByClause:
			fieldsValue := value.([]any)
			fields := make([]string, len(fieldsValue))
			for i, v := range fieldsValue {
				fields[i] = v.(string)
			}
			slct.GroupBy = immutable.Some(request.GroupBy{
				Fields: fields,
			})
		case request.ShowDeleted:
			slct.ShowDeleted = value.(bool)
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

	slct.Fields, err = parseSelectFields(exe, fieldObject, field.SelectionSet)
	if err != nil {
		return nil, err
	}

	return slct, err
}

func parseAggregate(
	exe *gql.ExecutionContext,
	parent *gql.Object,
	field *ast.Field,
) (*request.Aggregate, error) {
	targets := make([]*request.AggregateTarget, len(field.Arguments))

	fieldDef := gql.GetFieldDef(exe.Schema, parent, field.Name.Value)
	arguments := gql.GetArgumentValues(fieldDef.Args, field.Arguments, exe.VariableValues)

	for i, argument := range field.Arguments {
		name := argument.Name.Value
		value := arguments[name]

		switch v := value.(type) {
		case string:
			targets[i] = &request.AggregateTarget{
				HostName: v,
			}
		case map[string]any:
			var childName string
			var filter immutable.Option[request.Filter]
			var limit immutable.Option[uint64]
			var offset immutable.Option[uint64]
			var order immutable.Option[request.OrderBy]

			for _, f := range argument.Value.(*ast.ObjectValue).Fields {
				switch f.Name.Value {
				case request.FieldName:
					childName = v[request.FieldName].(string)

				case request.FilterClause:
					filter = immutable.Some(request.Filter{
						Conditions: v[request.FilterClause].(map[string]any),
					})

				case request.LimitClause:
					limit = immutable.Some(uint64(v[request.LimitClause].(int32)))

				case request.OffsetClause:
					offset = immutable.Some(uint64(v[request.OffsetClause].(int32)))

				case request.OrderClause:
					switch conditionsAST := f.Value.(type) {
					case *ast.EnumValue:
						// For inline arrays the order arg will be a simple enum declaring the order direction
						var orderDirection request.OrderDirection
						switch v[request.OrderClause].(int) {
						case 0:
							orderDirection = request.ASC

						case 1:
							orderDirection = request.DESC

						default:
							return nil, ErrInvalidOrderDirection
						}

						order = immutable.Some(request.OrderBy{
							Conditions: []request.OrderCondition{{
								Direction: orderDirection,
							}},
						})

					case *ast.ObjectValue:
						// For relations the order arg will be the complex order object as used by the host object
						// for non-aggregate ordering
						conditionsValue := v[request.OrderClause].(map[string]any)
						conditions, err := ParseConditionsInOrder(conditionsAST, conditionsValue)
						if err != nil {
							return nil, err
						}
						order = immutable.Some(request.OrderBy{
							Conditions: conditions,
						})
					}
				}
			}

			targets[i] = &request.AggregateTarget{
				HostName:  name,
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
