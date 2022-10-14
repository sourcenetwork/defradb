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
	"github.com/sourcenetwork/defradb/errors"
	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"
)

var dbAPIQueryNames = map[string]bool{
	"latestCommits": true,
	"commits":       true,
	"commit":        true,
}

type Query struct {
	Queries   []*OperationDefinition
	Mutations []*OperationDefinition
	Statement *ast.Document
}

type OperationDefinition struct {
	Name       string
	Selections []Selection
	Statement  *ast.OperationDefinition
	IsExplain  bool
}

func (q OperationDefinition) GetStatement() ast.Node {
	return q.Statement
}

type Selection interface {
	GetRoot() parserTypes.SelectionType
}

// Select is a complex Field with strong typing
// It used for sub types in a query. Includes
// fields, and query arguments like filters,
// limits, etc.
type Select struct {
	Name string
	// The identifier to be used in the rendered results, typically specified by
	// the user.
	Alias string

	DocKeys parserTypes.OptionalDocKeys
	CID     client.Option[string]

	// QueryType indicates what kind of query this is
	// Currently supports: ScanQuery, VersionedScanQuery
	QueryType parserTypes.SelectQueryType

	// Root is the top level query parsed type
	Root parserTypes.SelectionType

	Limit *parserTypes.Limit

	OrderBy *parserTypes.OrderBy

	GroupBy *parserTypes.GroupBy

	Filter *Filter

	Fields []Selection

	// Raw graphql statement
	Statement *ast.Field
}

func (s Select) GetRoot() parserTypes.SelectionType {
	return s.Root
}

// Field implements Selection
type Field struct {
	Name  string
	Alias string

	Root parserTypes.SelectionType
}

func (c Field) GetRoot() parserTypes.SelectionType {
	return c.Root
}

// ParseQuery parses a root ast.Document, and returns a
// formatted Query object.
// Requires a non-nil doc, will error if given a nil doc.
func ParseQuery(doc *ast.Document) (*Query, error) {
	if doc == nil {
		return nil, errors.New("ParseQuery requires a non-nil ast.Document")
	}
	q := &Query{
		Statement: doc,
		Queries:   make([]*OperationDefinition, 0),
		Mutations: make([]*OperationDefinition, 0),
	}

	for _, def := range q.Statement.Definitions {
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
					return nil, err
				}
				q.Mutations = append(q.Mutations, mdef)
			} else {
				return nil, errors.New("Unknown GraphQL operation type")
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
		if directive.Name.Value == parserTypes.ExplainLabel {
			return true
		}
	}

	return false
}

// parseQueryOperationDefinition parses the individual GraphQL
// 'query' operations, which there may be multiple of.
func parseQueryOperationDefinition(def *ast.OperationDefinition) (*OperationDefinition, error) {
	qdef := &OperationDefinition{
		Statement:  def,
		Selections: make([]Selection, len(def.SelectionSet.Selections)),
	}

	if def.Name != nil {
		qdef.Name = def.Name.Value
	}

	qdef.IsExplain = parseExplainDirective(def.Directives)

	for i, selection := range qdef.Statement.SelectionSet.Selections {
		var parsed Selection
		var err error
		switch node := selection.(type) {
		case *ast.Field:
			// which query type is this database API query object query etc.
			_, exists := dbAPIQueryNames[node.Name.Value]
			if exists {
				// the query matches a reserved DB API query name
				parsed, err = parseAPIQuery(node)
			} else {
				// the query doesn't match a reserve name
				// so its probably a generated query
				parsed, err = parseSelect(parserTypes.ObjectSelection, node, i)
			}
			if err != nil {
				return nil, err
			}

			qdef.Selections[i] = parsed
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
func parseSelect(rootType parserTypes.SelectionType, field *ast.Field, index int) (*Select, error) {
	slct := &Select{
		Alias:     getFieldAlias(field),
		Name:      field.Name.Value,
		Root:      rootType,
		Statement: field,
	}

	// parse arguments
	for _, argument := range field.Arguments {
		prop := argument.Name.Value
		astValue := argument.Value

		// parse filter
		if prop == parserTypes.FilterClause {
			obj := astValue.(*ast.ObjectValue)
			filter, err := NewFilter(obj)
			if err != nil {
				return slct, err
			}

			slct.Filter = filter
		} else if prop == parserTypes.DocKey { // parse single dockey query field
			val := astValue.(*ast.StringValue)
			slct.DocKeys = parserTypes.OptionalDocKeys{
				HasValue: true,
				Value:    []string{val.Value},
			}
		} else if prop == parserTypes.DocKeys {
			docKeyValues := astValue.(*ast.ListValue).Values
			docKeys := make([]string, len(docKeyValues))
			for i, value := range docKeyValues {
				docKeys[i] = value.(*ast.StringValue).Value
			}
			slct.DocKeys = parserTypes.OptionalDocKeys{
				HasValue: true,
				Value:    docKeys,
			}
		} else if prop == parserTypes.Cid { // parse single CID query field
			val := astValue.(*ast.StringValue)
			slct.CID = client.Some(val.Value)
		} else if prop == parserTypes.LimitClause { // parse limit/offset
			val := astValue.(*ast.IntValue)
			i, err := strconv.ParseInt(val.Value, 10, 64)
			if err != nil {
				return slct, err
			}
			if slct.Limit == nil {
				slct.Limit = &parserTypes.Limit{}
			}
			slct.Limit.Limit = i
		} else if prop == parserTypes.OffsetClause { // parse limit/offset
			val := astValue.(*ast.IntValue)
			i, err := strconv.ParseInt(val.Value, 10, 64)
			if err != nil {
				return slct, err
			}
			if slct.Limit == nil {
				slct.Limit = &parserTypes.Limit{}
			}
			slct.Limit.Offset = i
		} else if prop == parserTypes.OrderClause { // parse order by
			obj := astValue.(*ast.ObjectValue)
			cond, err := ParseConditionsInOrder(obj)
			if err != nil {
				return nil, err
			}
			slct.OrderBy = &parserTypes.OrderBy{
				Conditions: cond,
				Statement:  obj,
			}
		} else if prop == parserTypes.GroupByClause {
			obj := astValue.(*ast.ListValue)
			fields := make([]string, 0)
			for _, v := range obj.Values {
				fields = append(fields, v.GetValue().(string))
			}

			slct.GroupBy = &parserTypes.GroupBy{
				Fields: fields,
			}
		}

		if len(slct.DocKeys.Value) != 0 && slct.CID.HasValue() {
			slct.QueryType = parserTypes.VersionedScanQuery
		} else {
			slct.QueryType = parserTypes.ScanQuery
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

func getFieldAlias(field *ast.Field) string {
	if field.Alias == nil {
		return field.Name.Value
	}
	return field.Alias.Value
}

func parseSelectFields(root parserTypes.SelectionType, fields *ast.SelectionSet) ([]Selection, error) {
	selections := make([]Selection, len(fields.Selections))
	// parse field selections
	for i, selection := range fields.Selections {
		switch node := selection.(type) {
		case *ast.Field:
			if _, isAggregate := parserTypes.Aggregates[node.Name.Value]; isAggregate {
				s, err := parseSelect(root, node, i)
				if err != nil {
					return nil, err
				}
				selections[i] = s
			} else if node.SelectionSet == nil { // regular field
				selections[i] = parseField(root, node)
			} else { // sub type with extra fields
				subroot := root
				switch node.Name.Value {
				case parserTypes.VersionFieldName:
					subroot = parserTypes.CommitSelection
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
func parseField(root parserTypes.SelectionType, field *ast.Field) *Field {
	return &Field{
		Root:  root,
		Name:  field.Name.Value,
		Alias: getFieldAlias(field),
	}
}

func parseAPIQuery(field *ast.Field) (Selection, error) {
	switch field.Name.Value {
	case "latestCommits", "commits":
		return parseCommitSelect(field)
	default:
		return nil, errors.New("Unknown query")
	}
}
