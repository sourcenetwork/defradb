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

	"github.com/graphql-go/graphql/language/ast"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
)

type Query struct {
	Queries   []*OperationDefinition
	Mutations []*OperationDefinition
}

type OperationDefinition struct {
	Selections []Selection
	IsExplain  bool
}

type Selection any

// Select is a complex Field with strong typing
// It used for sub types in a query. Includes
// fields, and query arguments like filters,
// limits, etc.
type Select struct {
	Field

	DocKeys client.Option[[]string]
	CID     client.Option[string]

	// Root is the top level query parsed type
	Root request.SelectionType

	Limit   client.Option[uint64]
	Offset  client.Option[uint64]
	OrderBy client.Option[request.OrderBy]
	GroupBy client.Option[request.GroupBy]
	Filter  client.Option[Filter]

	Fields []Selection
}

func (s Select) validate() []error {
	result := []error{}

	result = append(result, s.validateShallow()...)

	for _, childSelection := range s.Fields {
		switch typedChildSelection := childSelection.(type) {
		case *Select:
			result = append(result, typedChildSelection.validateShallow()...)
		default:
			// Do nothing
		}
	}

	return result
}

func (s Select) validateShallow() []error {
	result := []error{}

	result = append(result, s.validateGroupBy()...)

	return result
}

func (s Select) validateGroupBy() []error {
	result := []error{}

	if !s.GroupBy.HasValue() {
		return result
	}

	for _, childSelection := range s.Fields {
		switch typedChildSelection := childSelection.(type) {
		case *Field:
			if typedChildSelection.Name == request.TypeNameFieldName {
				// _typeName is permitted
				continue
			}

			var fieldExistsInGroupBy bool
			for _, groupByField := range s.GroupBy.Value().Fields {
				if typedChildSelection.Name == groupByField {
					fieldExistsInGroupBy = true
					break
				}
			}
			if !fieldExistsInGroupBy {
				result = append(result, client.NewErrSelectOfNonGroupField(typedChildSelection.Name))
			}
		default:
			// Do nothing
		}
	}

	return result
}

// Field implements Selection
type Field struct {
	Name  string
	Alias client.Option[string]
}

// ParseQuery parses a root ast.Document, and returns a
// formatted Query object.
// Requires a non-nil doc, will error if given a nil doc.
func ParseQuery(doc *ast.Document) (*Query, []error) {
	if doc == nil {
		return nil, []error{errors.New("parseQuery requires a non-nil ast.Document")}
	}
	q := &Query{
		Queries:   make([]*OperationDefinition, 0),
		Mutations: make([]*OperationDefinition, 0),
	}

	for _, def := range doc.Definitions {
		switch node := def.(type) {
		case *ast.OperationDefinition:
			if node.Operation == "query" {
				// parse query operation definition.
				qdef, err := parseQueryOperationDefinition(node)
				if err != nil {
					return nil, err
				}
				q.Queries = append(q.Queries, qdef)
			} else if node.Operation == "mutation" {
				// parse mutation operation definition.
				mdef, err := parseMutationOperationDefinition(node)
				if err != nil {
					return nil, []error{err}
				}
				q.Mutations = append(q.Mutations, mdef)
			} else {
				return nil, []error{errors.New("unknown GraphQL operation type")}
			}
		}
	}

	return q, nil
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
func parseQueryOperationDefinition(def *ast.OperationDefinition) (*OperationDefinition, []error) {
	qdef := &OperationDefinition{
		Selections: make([]Selection, len(def.SelectionSet.Selections)),
	}

	qdef.IsExplain = parseExplainDirective(def.Directives)

	for i, selection := range def.SelectionSet.Selections {
		var parsedSelection Selection
		switch node := selection.(type) {
		case *ast.Field:
			if _, isCommitQuery := request.CommitQueries[node.Name.Value]; isCommitQuery {
				parsed, err := parseCommitSelect(node)
				if err != nil {
					return nil, []error{err}
				}

				parsedSelection = parsed
			} else if _, isAggregate := request.Aggregates[node.Name.Value]; isAggregate {
				parsed, err := parseAggregate(node, i)
				if err != nil {
					return nil, []error{err}
				}

				// Top-level aggregates must be wrapped in a top-level Select for now
				parsedSelection = &Select{
					Field: Field{
						Name:  parsed.Name,
						Alias: parsed.Alias,
					},
					Fields: []Selection{
						parsed,
					},
				}
			} else {
				// the query doesn't match a reserve name
				// so its probably a generated query
				parsed, err := parseSelect(request.ObjectSelection, node, i)
				if err != nil {
					return nil, []error{err}
				}

				errors := parsed.validate()
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
func parseSelect(rootType request.SelectionType, field *ast.Field, index int) (*Select, error) {
	slct := &Select{
		Field: Field{
			Name:  field.Name.Value,
			Alias: getFieldAlias(field),
		},
		Root: rootType,
	}

	// parse arguments
	for _, argument := range field.Arguments {
		prop := argument.Name.Value
		astValue := argument.Value

		// parse filter
		if prop == request.FilterClause {
			obj := astValue.(*ast.ObjectValue)
			filter, err := NewFilter(obj)
			if err != nil {
				return slct, err
			}

			slct.Filter = filter
		} else if prop == request.DocKey { // parse single dockey query field
			val := astValue.(*ast.StringValue)
			slct.DocKeys = client.Some([]string{val.Value})
		} else if prop == request.DocKeys {
			docKeyValues := astValue.(*ast.ListValue).Values
			docKeys := make([]string, len(docKeyValues))
			for i, value := range docKeyValues {
				docKeys[i] = value.(*ast.StringValue).Value
			}
			slct.DocKeys = client.Some(docKeys)
		} else if prop == request.Cid { // parse single CID query field
			val := astValue.(*ast.StringValue)
			slct.CID = client.Some(val.Value)
		} else if prop == request.LimitClause { // parse limit/offset
			val := astValue.(*ast.IntValue)
			limit, err := strconv.ParseUint(val.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			slct.Limit = client.Some(limit)
		} else if prop == request.OffsetClause { // parse limit/offset
			val := astValue.(*ast.IntValue)
			offset, err := strconv.ParseUint(val.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			slct.Offset = client.Some(offset)
		} else if prop == request.OrderClause { // parse order by
			obj := astValue.(*ast.ObjectValue)
			cond, err := ParseConditionsInOrder(obj)
			if err != nil {
				return nil, err
			}
			slct.OrderBy = client.Some(
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

			slct.GroupBy = client.Some(
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
	var err error
	slct.Fields, err = parseSelectFields(slct.Root, field.SelectionSet)
	if err != nil {
		return nil, err
	}

	return slct, err
}

func getFieldAlias(field *ast.Field) client.Option[string] {
	if field.Alias == nil {
		return client.None[string]()
	}
	return client.Some(field.Alias.Value)
}

func parseSelectFields(root request.SelectionType, fields *ast.SelectionSet) ([]Selection, error) {
	selections := make([]Selection, len(fields.Selections))
	// parse field selections
	for i, selection := range fields.Selections {
		switch node := selection.(type) {
		case *ast.Field:
			if _, isAggregate := request.Aggregates[node.Name.Value]; isAggregate {
				s, err := parseAggregate(node, i)
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
				s, err := parseSelect(subroot, node, i)
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
func parseField(field *ast.Field) *Field {
	return &Field{
		Name:  field.Name.Value,
		Alias: getFieldAlias(field),
	}
}

type Aggregate struct {
	Field

	Targets []*AggregateTarget
}

type AggregateTarget struct {
	HostName  string
	ChildName client.Option[string]

	Limit   client.Option[uint64]
	Offset  client.Option[uint64]
	OrderBy client.Option[request.OrderBy]
	Filter  client.Option[Filter]
}

func parseAggregate(field *ast.Field, index int) (*Aggregate, error) {
	targets := make([]*AggregateTarget, len(field.Arguments))

	for i, argument := range field.Arguments {
		switch argumentValue := argument.Value.GetValue().(type) {
		case string:
			targets[i] = &AggregateTarget{
				HostName: argumentValue,
			}
		case []*ast.ObjectField:
			hostName := argument.Name.Value
			var childName string
			var filter client.Option[Filter]
			var limit client.Option[uint64]
			var offset client.Option[uint64]
			var order client.Option[request.OrderBy]

			fieldArg, hasFieldArg := tryGet(argumentValue, request.Field)
			if hasFieldArg {
				if innerPathStringValue, isString := fieldArg.Value.GetValue().(string); isString {
					childName = innerPathStringValue
				}
			}

			filterArg, hasFilterArg := tryGet(argumentValue, request.FilterClause)
			if hasFilterArg {
				filterValue, err := NewFilter(filterArg.Value.(*ast.ObjectValue))
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
				limit = client.Some(limitValue)
			}

			offsetArg, hasOffsetArg := tryGet(argumentValue, request.OffsetClause)
			if hasOffsetArg {
				offsetValue, err := strconv.ParseUint(offsetArg.Value.(*ast.IntValue).Value, 10, 64)
				if err != nil {
					return nil, err
				}
				offset = client.Some(offsetValue)
			}

			orderArg, hasOrderArg := tryGet(argumentValue, request.OrderClause)
			if hasOrderArg {
				switch orderArgValue := orderArg.Value.(type) {
				case *ast.EnumValue:
					// For inline arrays the order arg will be a simple enum declaring the order direction
					orderDirectionString := orderArgValue.Value
					orderDirection := request.OrderDirection(orderDirectionString)

					order = client.Some(
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

					order = client.Some(
						request.OrderBy{
							Conditions: orderConditions,
						},
					)
				}
			}

			targets[i] = &AggregateTarget{
				HostName:  hostName,
				ChildName: client.Some(childName),
				Filter:    filter,
				Limit:     limit,
				Offset:    offset,
				OrderBy:   order,
			}
		}
	}

	return &Aggregate{
		Field: Field{
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
