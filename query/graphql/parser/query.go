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
	"errors"
	"fmt"
	"strconv"

	"github.com/graphql-go/graphql/language/ast"
)

type SelectionType int

const (
	NoneSelection = iota
	ObjectSelection
	CommitSelection

	VersionFieldName = "_version"
	GroupFieldName   = "_group"
	DocKeyFieldName  = "_key"
	CountFieldName   = "_count"
	SumFieldName     = "_sum"
	HiddenFieldName  = "_hidden"
)

// Enum for different types of read Select queries
type SelectQueryType int

const (
	ScanQuery = iota
	VersionedScanQuery
)

var dbAPIQueryNames = map[string]bool{
	"latestCommits": true,
	"allCommits":    true,
	"commit":        true,
}

var ReservedFields = map[string]bool{
	VersionFieldName: true,
	GroupFieldName:   true,
	CountFieldName:   true,
	SumFieldName:     true,
	HiddenFieldName:  true,
	DocKeyFieldName:  true,
}

var Aggregates = map[string]struct{}{
	CountFieldName: {},
	SumFieldName:   {},
}

type Query struct {
	Queries   []*OperationDefinition
	Mutations []*OperationDefinition
	Statement *ast.Document
}

func (q Query) GetStatement() ast.Node {
	return q.Statement
}

type OperationDefinition struct {
	Name       string
	Selections []Selection

	Statement *ast.OperationDefinition
}

func (q OperationDefinition) GetStatement() ast.Node {
	return q.Statement
}

// type SelectionSet struct {
// 	Selections []Selection
// }

type Selection interface {
	Statement
	GetName() string
	GetAlias() string
	GetSelections() []Selection
	GetRoot() SelectionType
}

// Select is a complex Field with strong typing
// It used for sub types in a query. Includes
// fields, and query arguments like filters,
// limits, etc.
type Select struct {
	Name           string
	Alias          string
	CollectionName string

	// QueryType indicates what kind of query this is
	// Currently supports: ScanQuery, VersionedScanQuery
	QueryType SelectQueryType

	// Root is the top level query parsed type
	Root SelectionType

	DocKeys []string
	CID     string

	Filter  *Filter
	Limit   *Limit
	OrderBy *OrderBy
	GroupBy *GroupBy

	Fields []Selection

	// raw graphql statement
	Statement *ast.Field
}

func (s Select) GetRoot() SelectionType {
	return s.Root
}

func (s Select) GetStatement() ast.Node {
	return s.Statement
}

func (s Select) GetSelections() []Selection {
	return s.Fields
}

func (s Select) GetName() string {
	return s.Name
}

func (s Select) GetAlias() string {
	return s.Alias
}

// Field implements Selection
type Field struct {
	Name  string
	Alias string

	Root SelectionType

	// raw graphql statement
	Statement *ast.Field
}

func (c Field) GetRoot() SelectionType {
	return c.Root
}

// GetSelectionSet implements Selection
func (f Field) GetSelections() []Selection {
	return []Selection{}
}

func (f Field) GetName() string {
	return f.Name
}

func (f Field) GetAlias() string {
	return f.Alias
}

func (f Field) GetStatement() ast.Node {
	return f.Statement
}

type GroupBy struct {
	Fields []string
}

type SortDirection string

const (
	ASC  SortDirection = "ASC"
	DESC SortDirection = "DESC"
)

var (
	NameToSortDirection = map[string]SortDirection{
		"ASC":  ASC,
		"DESC": DESC,
	}
)

type SortCondition struct {
	// field may be a compound field statement
	// since the sort statement allows sorting on
	// sub objects.
	//
	// Given the statement: {sort: {author: {birthday: DESC}}}
	// The field value would be "author.birthday"
	// and the direction would be "DESC"
	Field     string
	Direction SortDirection
}
type OrderBy struct {
	Conditions []SortCondition
	Statement  *ast.ObjectValue
}

type Limit struct {
	Limit  int64
	Offset int64
}

// type SubQuery struct{}

// type

// ParseQuery parses a root ast.Document, and returns a
// formatted Query object.
// Requires a non-nil doc, will error if given a nil
// doc
func ParseQuery(doc *ast.Document) (*Query, error) {
	if doc == nil {
		return nil, errors.New("ParseQuery requires a non nil ast.Document")
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
				// parse query or mutation operation definition
				qdef, err := parseQueryOperationDefinition(node)
				if err != nil {
					return nil, err
				}
				q.Queries = append(q.Queries, qdef)
			} else if node.Operation == "mutation" {
				mdef, err := parseMutationOperationDefinition(node)
				if err != nil {
					return nil, err
				}
				q.Mutations = append(q.Mutations, mdef)
			} else {
				return nil, errors.New("Unkown graphql operation type")
			}
		}
	}

	return q, nil
}

// parseOperationDefinition parses the individual GraphQL
// 'query' operations, which there may be multiple of.
func parseQueryOperationDefinition(def *ast.OperationDefinition) (*OperationDefinition, error) {
	qdef := &OperationDefinition{
		Statement:  def,
		Selections: make([]Selection, len(def.SelectionSet.Selections)),
	}
	if def.Name != nil {
		qdef.Name = def.Name.Value
	}
	for i, selection := range qdef.Statement.SelectionSet.Selections {
		var parsed Selection
		var err error
		switch node := selection.(type) {
		case *ast.Field:
			// which query type is this
			// database API query
			// object query
			// etc
			_, exists := dbAPIQueryNames[node.Name.Value]
			if exists {
				// the query matches a reserved DB API query name
				parsed, err = parseAPIQuery(node)
			} else {
				// the query doesn't match a reserve name
				// so its probably a generated query
				parsed, err = parseSelect(ObjectSelection, node)
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
func parseSelect(rootType SelectionType, field *ast.Field) (*Select, error) {
	slct := &Select{
		Root:      rootType,
		Statement: field,
	}
	slct.Name = field.Name.Value
	if field.Alias != nil {
		slct.Alias = field.Alias.Value
	}

	// parse arguments
	for _, argument := range field.Arguments {
		prop := argument.Name.Value
		// parse filter
		if prop == "filter" {
			obj := argument.Value.(*ast.ObjectValue)
			filter, err := NewFilter(obj)
			if err != nil {
				return slct, err
			}

			slct.Filter = filter
		} else if prop == "dockey" { // parse single dockey query field
			val := argument.Value.(*ast.StringValue)
			slct.DocKeys = []string{val.Value}
		} else if prop == "dockeys" {
			docKeyValues := argument.Value.(*ast.ListValue).Values
			docKeys := make([]string, len(docKeyValues))
			for i, value := range docKeyValues {
				docKeys[i] = value.(*ast.StringValue).Value
			}
			slct.DocKeys = docKeys
		} else if prop == "cid" { // parse single CID query field
			val := argument.Value.(*ast.StringValue)
			slct.CID = val.Value
		} else if prop == "limit" { // parse limit/offset
			val := argument.Value.(*ast.IntValue)
			i, err := strconv.ParseInt(val.Value, 10, 64)
			if err != nil {
				return slct, err
			}
			if slct.Limit == nil {
				slct.Limit = &Limit{}
			}
			slct.Limit.Limit = i
		} else if prop == "offset" { // parse limit/offset
			val := argument.Value.(*ast.IntValue)
			i, err := strconv.ParseInt(val.Value, 10, 64)
			if err != nil {
				return slct, err
			}
			if slct.Limit == nil {
				slct.Limit = &Limit{}
			}
			slct.Limit.Offset = i
		} else if prop == "order" { // parse sort (order by)
			obj := argument.Value.(*ast.ObjectValue)
			cond, err := ParseConditionsInOrder(obj)
			if err != nil {
				return nil, err
			}
			slct.OrderBy = &OrderBy{
				Conditions: cond,
				Statement:  obj,
			}
		} else if prop == "groupBy" {
			obj := argument.Value.(*ast.ListValue)
			fields := make([]string, 0)
			for _, v := range obj.Values {
				fields = append(fields, v.GetValue().(string))
			}

			slct.GroupBy = &GroupBy{
				Fields: fields,
			}
		}

		if len(slct.DocKeys) != 0 && len(slct.CID) != 0 {
			slct.QueryType = VersionedScanQuery
		} else {
			slct.QueryType = ScanQuery
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

func parseSelectFields(root SelectionType, fields *ast.SelectionSet) ([]Selection, error) {
	selections := make([]Selection, len(fields.Selections))
	// parse field selections
	for i, selection := range fields.Selections {
		switch node := selection.(type) {
		case *ast.Field:
			if node.SelectionSet == nil { // regular field
				f, err := parseField(i, root, node)
				if err != nil {
					return nil, err
				}
				selections[i] = f
			} else { // sub type with extra fields
				subroot := root
				switch node.Name.Value {
				case "_version":
					subroot = CommitSelection
				}
				s, err := parseSelect(subroot, node)
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
func parseField(i int, root SelectionType, field *ast.Field) (*Field, error) {
	var name string
	var alias string

	if _, isAggregate := Aggregates[field.Name.Value]; isAggregate {
		name = fmt.Sprintf("_agg%v", i)
		if field.Alias == nil {
			alias = field.Name.Value
		} else {
			alias = field.Alias.Value
		}
	} else {
		name = field.Name.Value
		if field.Alias != nil {
			alias = field.Alias.Value
		}
	}

	f := &Field{
		Root:      root,
		Name:      name,
		Statement: field,
		Alias:     alias,
	}
	return f, nil
}

func parseAPIQuery(field *ast.Field) (Selection, error) {
	switch field.Name.Value {
	case "latestCommits", "allCommits", "commit":
		return parseCommitSelect(field)
	default:
		return nil, errors.New("Unknown query")
	}
}

// Returns the source of the aggregate as requested by the consumer
func (field Field) GetAggregateSource() ([]string, error) {

	if len(field.Statement.Arguments) == 0 {
		return []string{}, fmt.Errorf(
			"Aggregate must be provided with a property to aggregate.",
		)
	}

	var path []string
	switch argumentValue := field.Statement.Arguments[0].Value.GetValue().(type) {
	case string:
		path = []string{argumentValue}
	case []*ast.ObjectField:
		if len(argumentValue) == 0 {
			return []string{}, fmt.Errorf(
				"Unexpected error: aggregate field contained no child field selector",
			)
		}
		innerPath := argumentValue[0].Value.GetValue()
		if innerPathStringValue, isString := innerPath.(string); isString {
			path = []string{argumentValue[0].Name.Value, innerPathStringValue}
		} else {
			// If the inner path is not a string, this must mean the field is an inline array
			//  in which case we only want the base path
			path = []string{argumentValue[0].Name.Value}
		}
	}

	return path, nil
}
