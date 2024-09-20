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

	for _, argument := range field.Arguments {
		name := argument.Name.Value
		value := arguments[name]

		switch name {
		case request.FilterClause:
			if v, ok := value.(map[string]any); ok {
				slct.Filter = immutable.Some(request.Filter{Conditions: v})
			}

		case request.DocIDArgName: // parse single DocID field
			if v, ok := value.(string); ok {
				slct.DocIDs = immutable.Some([]string{v})
			}

		case request.DocIDsArgName:
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}
			docIDs := make([]string, len(v))
			for i, value := range v {
				docIDs[i] = value.(string)
			}
			slct.DocIDs = immutable.Some(docIDs)

		case request.Cid: // parse single CID query field
			if v, ok := value.(string); ok {
				slct.CID = immutable.Some(v)
			}

		case request.LimitClause: // parse limit/offset
			if v, ok := value.(int32); ok {
				slct.Limit = immutable.Some(uint64(v))
			}

		case request.OffsetClause: // parse limit/offset
			if v, ok := value.(int32); ok {
				slct.Offset = immutable.Some(uint64(v))
			}

		case request.OrderClause: // parse order by
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}
			conditions, err := parseOrderConditionList(v)
			if err != nil {
				return nil, err
			}
			slct.OrderBy = immutable.Some(request.OrderBy{
				Conditions: conditions,
			})

		case request.GroupByClause:
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}
			fields := make([]string, len(v))
			for i, c := range v {
				fields[i] = c.(string)
			}
			slct.GroupBy = immutable.Some(request.GroupBy{
				Fields: fields,
			})

		case request.ShowDeleted:
			if v, ok := value.(bool); ok {
				slct.ShowDeleted = v
			}
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
	fieldDef := gql.GetFieldDef(exe.Schema, parent, field.Name.Value)
	arguments := gql.GetArgumentValues(fieldDef.Args, field.Arguments, exe.VariableValues)

	var targets []*request.AggregateTarget
	for _, argument := range field.Arguments {
		name := argument.Name.Value

		switch v := arguments[name].(type) {
		case string:
			targets = append(targets, &request.AggregateTarget{
				HostName: v,
			})

		case map[string]any:
			target, err := parseAggregateTarget(name, v)
			if err != nil {
				return nil, err
			}
			targets = append(targets, target)
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

func parseAggregateTarget(
	hostName string,
	arguments map[string]any,
) (*request.AggregateTarget, error) {
	var childName string
	var filter immutable.Option[request.Filter]
	var limit immutable.Option[uint64]
	var offset immutable.Option[uint64]
	var order immutable.Option[request.OrderBy]

	for name, value := range arguments {
		switch name {
		case request.FieldName:
			if v, ok := value.(string); ok {
				childName = v
			}

		case request.FilterClause:
			if v, ok := value.(map[string]any); ok {
				filter = immutable.Some(request.Filter{Conditions: v})
			}

		case request.LimitClause:
			if v, ok := value.(int32); ok {
				limit = immutable.Some(uint64(v))
			}

		case request.OffsetClause:
			if v, ok := value.(int32); ok {
				offset = immutable.Some(uint64(v))
			}

		case request.OrderClause:
			switch t := value.(type) {
			case int:
				// For inline arrays the order arg will be a simple enum declaring the order direction
				dir, err := parseOrderDirection(t)
				if err != nil {
					return nil, err
				}
				order = immutable.Some(request.OrderBy{
					Conditions: []request.OrderCondition{{Direction: dir}},
				})

			case []any:
				// For relations the order arg will be the complex order object as used by the host object
				// for non-aggregate ordering
				conditions, err := parseOrderConditionList(t)
				if err != nil {
					return nil, err
				}
				order = immutable.Some(request.OrderBy{
					Conditions: conditions,
				})
			}
		}
	}

	return &request.AggregateTarget{
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
	}, nil
}
