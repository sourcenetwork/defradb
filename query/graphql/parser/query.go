package parser

import (
	"errors"
	"strconv"

	"github.com/graphql-go/graphql/language/ast"
)

type Query struct {
	Queries   []*QueryDefinition
	Statement *ast.Document
}

func (q Query) GetStatement() ast.Node {
	return q.Statement
}

type QueryDefinition struct {
	Name       string
	Selections []Selection

	Statement *ast.OperationDefinition
}

func (q QueryDefinition) GetStatement() ast.Node {
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
}

// type

// Select is a complex Field with strong typing
// It used for sub types in a query. Includes
// fields, and query arguments like filters,
// limits, etc.
type Select struct {
	Name  string
	Alias string

	Filter  *Filter
	Limit   *Limit
	OrderBy *OrderBy
	GroupBy *GroupBy

	Fields []Selection

	// raw graphql statement
	Statement *ast.Field
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

	// raw graphql statement
	Statement *ast.Field
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

type GroupBy struct{}
type OrderBy struct {
	Conditions map[string]interface{}
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
		Queries:   make([]*QueryDefinition, len(doc.Definitions)),
	}
	for i, def := range q.Statement.Definitions {
		switch node := def.(type) {
		case *ast.OperationDefinition:
			qdef, err := parseOperationDefinition(node)
			if err != nil {
				return nil, err
			}
			q.Queries[i] = qdef
		}
	}

	return q, nil
}

// parseOperationDefintition parses the individual GraphQL
// 'query' operations, which there may be mulitple of.
func parseOperationDefinition(def *ast.OperationDefinition) (*QueryDefinition, error) {
	qdef := &QueryDefinition{
		Statement:  def,
		Selections: make([]Selection, len(def.SelectionSet.Selections)),
	}
	if def.Name != nil {
		qdef.Name = def.Name.Value
	}
	for i, selection := range qdef.Statement.SelectionSet.Selections {
		switch node := selection.(type) {
		case *ast.Field:
			slct, err := parseSelect(node)
			if err != nil {
				return nil, err
			}

			qdef.Selections[i] = slct
		}
	}
	return qdef, nil
}

// parseSelect parses a typed selection field
// which includes sub fields, and may include
// filters, limits, orders, etc..
func parseSelect(field *ast.Field) (*Select, error) {
	slct := &Select{Statement: field}
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
		} else if prop == "limit" { // parse limit/offset
			val := argument.Value.(*ast.IntValue)
			i, err := strconv.ParseInt(val.Value, 10, 64)
			if err != nil {
				return slct, err
			}
			slct.Limit.Limit = i
		} else if prop == "offset" {
			val := argument.Value.(*ast.IntValue)
			i, err := strconv.ParseInt(val.Value, 10, 64)
			if err != nil {
				return slct, err
			}
			slct.Limit.Offset = i
		}

		// @todo: parse orderby

		// @todo: parse groupby
	}

	// parse field selections
	for _, selection := range field.SelectionSet.Selections {
		switch node := selection.(type) {
		case *ast.Field:
			if node.SelectionSet == nil { // regular field
				f := parseField(node)
				slct.Fields = append(slct.Fields, f)
			} else { // sub type with extra fields
				s, err := parseSelect(node)
				if err != nil {
					return slct, err
				}
				slct.Fields = append(slct.Fields, s)
			}
		}
	}

	return slct, nil
}

// parseField simply parses the Name/Alias
// into a Field type
func parseField(field *ast.Field) *Field {
	f := &Field{Name: field.Name.Value}
	if field.Alias != nil {
		f.Alias = field.Alias.Value
	}
	return f
}
