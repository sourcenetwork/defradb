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
	"encoding/json"
	"errors"
	"strings"

	"github.com/graphql-go/graphql/language/ast"

	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"
)

type MutationType int

const (
	NoneMutationType = MutationType(iota)
	CreateObjects
	UpdateObjects
	DeleteObjects
)

var (
	mutationNameToType = map[string]MutationType{
		"create": CreateObjects,
		"update": UpdateObjects,
		"delete": DeleteObjects,
	}
)

var (
	ErrEmptyDataPayload = errors.New("given data payload is empty")
)

type ObjectPayload struct {
	Object map[string]interface{}
	Array  []interface{}
}

// NewObjectPayload parses a given payload string as JSON
// and returns a ObjectPayload struct decoded with either
// a JSON object, or JSON array.
func NewObjectPayload(payload string) (ObjectPayload, error) {
	obj := ObjectPayload{}
	if payload == "" {
		return obj, errors.New("Object payload value cannot be empty")
	}
	var d interface{}
	err := json.Unmarshal([]byte(payload), &d)
	if err != nil {
		return obj, err
	}

	switch v := d.(type) {

	// array usually means its a JSON PATCH object, unless its a create, then its
	//  just multiple documents
	case []interface{}:
		obj.Array = v

	case map[string]interface{}:
		obj.Object = v

	default:
		return obj, errors.New(
			"Object payload value has unknown structure, must be a JSON object or array",
		)
	}

	return obj, nil
}

// Mutation is a field on the MutationType
// of a graphql query. It includes all the possible
// arguments and all
//
// @todo: Change name to ObjectMutation to indicate
// generated object mutation actions
type Mutation struct {
	Name  string
	Alias string
	Type  MutationType

	// Schema is the target schema/collection
	// if this mutation is on an object.
	Schema string

	IDs    []string
	Filter *Filter
	Data   string

	Fields []Selection

	Statement *ast.Field
}

func (m Mutation) GetRoot() parserTypes.SelectionType {
	return parserTypes.ObjectSelection
}

func (m Mutation) GetStatement() ast.Node {
	return m.Statement
}

func (m Mutation) GetSelections() []Selection {
	return m.Fields
}

func (m Mutation) GetName() string {
	return m.Name
}

func (m Mutation) GetAlias() string {
	return m.Alias
}

// ToSelect returns a basic Select object, with the same Name,
// Alias, and Fields as the Mutation object. Used to create a
// Select planNode for the mutation return objects
func (m Mutation) ToSelect() *Select {
	return &Select{
		Name:         m.Schema,
		ExternalName: m.Schema,
		Alias:        m.Alias,
		Fields:       m.Fields,
		DocKeys:      m.IDs,
		Filter:       m.Filter,
	}
}

// parseMutationOperationDefinition parses the individual GraphQL
// 'mutation' operations, which there may be multiple of.
func parseMutationOperationDefinition(def *ast.OperationDefinition) (*OperationDefinition, error) {

	qdef := &OperationDefinition{
		Statement:  def,
		Selections: make([]Selection, len(def.SelectionSet.Selections)),
	}

	if def.Name != nil {
		qdef.Name = def.Name.Value
	}

	qdef.IsExplain = parseExplainDirective(def.Directives)

	for i, selection := range qdef.Statement.SelectionSet.Selections {
		switch node := selection.(type) {
		case *ast.Field:
			mut, err := parseMutation(node)
			if err != nil {
				return nil, err
			}

			qdef.Selections[i] = mut
		}
	}
	return qdef, nil
}

// @todo: Create separate mutation parse functions
// for generated object mutations, and general
// API mutations.

// parseMutation parses a typed mutation field
// which includes sub fields, and may include
// filters, IDs, payloads, etc.
func parseMutation(field *ast.Field) (*Mutation, error) {
	mut := &Mutation{Statement: field}
	mut.Name = field.Name.Value
	if field.Alias != nil {
		mut.Alias = field.Alias.Value
	}

	// parse the mutation type
	// mutation names are either generated from a type
	// which means they are in the form name_type, where
	// the name is the object mutation name (ie: create, update, delete)
	// or its an general API mutation, which is in the form
	// name (camelCase).
	// This means we can split on the "_" character, and always
	// get back one of our defined types.
	mutNameParts := strings.Split(mut.Name, "_")
	typeStr := mutNameParts[0]
	var ok bool
	mut.Type, ok = mutationNameToType[typeStr]
	if !ok {
		return nil, errors.New("unknown mutation name")
	}

	if len(mutNameParts) > 1 { // only generated object mutations
		// reconstruct the name.
		// if the schema/collection name is eg: my_book
		// then the mutation name would be create_my_book
		// so we need to recreate the string my_book, which
		// has been split by "_", so we just join by "_"
		mut.Schema = strings.Join(mutNameParts[1:], "_")
	}

	// parse arguments
	for _, argument := range field.Arguments {
		prop := argument.Name.Value
		// parse each individual arg type seperately
		if prop == parserTypes.Data { // parse data
			raw := argument.Value.(*ast.StringValue)
			if raw.Value == "" {
				return nil, ErrEmptyDataPayload
			}
			mut.Data = raw.Value
		} else if prop == parserTypes.FilterClause { // parse filter
			obj := argument.Value.(*ast.ObjectValue)
			filter, err := NewFilter(obj)
			if err != nil {
				return nil, err
			}

			mut.Filter = filter
		} else if prop == parserTypes.Id {
			raw := argument.Value.(*ast.StringValue)
			mut.IDs = []string{raw.Value}
		} else if prop == parserTypes.Ids {
			raw := argument.Value.(*ast.ListValue)
			mut.IDs = make([]string, len(raw.Values))
			for i, val := range raw.Values {
				id, ok := val.(*ast.StringValue)
				if !ok {
					return nil, errors.New("ids argument has a non string value")
				}
				mut.IDs[i] = id.Value
			}

		}
	}

	// if theres no field selections, just return
	if field.SelectionSet == nil {
		return mut, nil
	}

	var err error
	mut.Fields, err = parseSelectFields(mut.GetRoot(), field.SelectionSet)
	return mut, err
}
