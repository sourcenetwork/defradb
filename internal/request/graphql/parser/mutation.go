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
	"strings"

	gql "github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

var (
	mutationNameToType = map[string]request.MutationType{
		"create": request.CreateObjects,
		"update": request.UpdateObjects,
		"delete": request.DeleteObjects,
	}
)

// parseMutationOperationDefinition parses the individual GraphQL
// 'mutation' operations, which there may be multiple of.
func parseMutationOperationDefinition(
	schema gql.Schema,
	def *ast.OperationDefinition,
) (*request.OperationDefinition, error) {
	qdef := &request.OperationDefinition{
		Selections: make([]request.Selection, len(def.SelectionSet.Selections)),
	}
	if def.Name != nil {
		qdef.Name = def.Name.Value
	}

	for i, selection := range def.SelectionSet.Selections {
		switch node := selection.(type) {
		case *ast.Field:
			mut, err := parseMutation(schema, schema.MutationType(), node)
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
func parseMutation(schema gql.Schema, parent *gql.Object, field *ast.Field) (*request.ObjectMutation, error) {
	mut := &request.ObjectMutation{
		Field: request.Field{
			Name:  field.Name.Value,
			Alias: getFieldAlias(field),
		},
	}

	fieldDef := gql.GetFieldDef(schema, parent, mut.Name)

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
		return nil, ErrUnknownMutationName
	}

	if len(mutNameParts) > 1 { // only generated object mutations
		// reconstruct the name.
		// if the schema/collection name is eg: my_book
		// then the mutation name would be create_my_book
		// so we need to recreate the string my_book, which
		// has been split by "_", so we just join by "_"
		mut.Collection = strings.Join(mutNameParts[1:], "_")
	}

	// parse arguments
	for _, argument := range field.Arguments {
		prop := argument.Name.Value
		// parse each individual arg type seperately
		if prop == request.Input { // parse input
			raw := argument.Value.(*ast.ObjectValue)
			mut.Input = parseMutationInputObject(raw)
		} else if prop == request.Inputs {
			raw := argument.Value.(*ast.ListValue)

			mut.Inputs = make([]map[string]any, len(raw.Values))

			for i, val := range raw.Values {
				doc, ok := val.(*ast.ObjectValue)
				if !ok {
					return nil, client.NewErrUnexpectedType[*ast.ObjectValue]("doc array element", val)
				}
				mut.Inputs[i] = parseMutationInputObject(doc)
			}
		} else if prop == request.FilterClause { // parse filter
			obj := argument.Value.(*ast.ObjectValue)
			filterType, ok := getArgumentType(fieldDef, request.FilterClause)
			if !ok {
				return nil, ErrFilterMissingArgumentType
			}
			filter, err := NewFilter(obj, filterType)
			if err != nil {
				return nil, err
			}

			mut.Filter = filter
		} else if prop == request.DocIDArgName {
			raw := argument.Value.(*ast.StringValue)
			mut.DocIDs = immutable.Some([]string{raw.Value})
		} else if prop == request.DocIDsArgName {
			raw := argument.Value.(*ast.ListValue)
			ids := make([]string, len(raw.Values))
			for i, val := range raw.Values {
				id, ok := val.(*ast.StringValue)
				if !ok {
					return nil, client.NewErrUnexpectedType[*ast.StringValue]("ids argument", val)
				}
				ids[i] = id.Value
			}
			mut.DocIDs = immutable.Some(ids)
		} else if prop == request.EncryptDocArgName {
			mut.Encrypt = argument.Value.(*ast.BooleanValue).Value
		} else if prop == request.EncryptFieldsArgName {
			raw := argument.Value.(*ast.ListValue)
			fieldNames := make([]string, len(raw.Values))
			for i, val := range raw.Values {
				fieldNames[i] = val.GetValue().(string)
			}
			mut.EncryptFields = fieldNames
		}
	}

	// if theres no field selections, just return
	if field.SelectionSet == nil {
		return mut, nil
	}

	fieldObject, err := typeFromFieldDef(fieldDef)
	if err != nil {
		return nil, err
	}

	mut.Fields, err = parseSelectFields(schema, fieldObject, field.SelectionSet)
	return mut, err
}

// parseMutationInput parses the correct underlying
// value type of the given ast.Value
func parseMutationInput(val ast.Value) any {
	switch t := val.(type) {
	case *ast.IntValue:
		return gql.Int.ParseLiteral(val)
	case *ast.FloatValue:
		return gql.Float.ParseLiteral(val)
	case *ast.BooleanValue:
		return t.Value
	case *ast.StringValue:
		return t.Value
	case *ast.ObjectValue:
		return parseMutationInputObject(t)
	case *ast.ListValue:
		return parseMutationInputList(t)
	default:
		return val.GetValue()
	}
}

// parseMutationInputList parses the correct underlying
// value type for all of the values in the ast.ListValue
func parseMutationInputList(val *ast.ListValue) []any {
	list := make([]any, 0)
	for _, val := range val.Values {
		list = append(list, parseMutationInput(val))
	}
	return list
}

// parseMutationInputObject parses the correct underlying
// value type for all of the fields in the ast.ObjectValue
func parseMutationInputObject(val *ast.ObjectValue) map[string]any {
	obj := make(map[string]any)
	for _, field := range val.Fields {
		obj[field.Name.Value] = parseMutationInput(field.Value)
	}
	return obj
}
