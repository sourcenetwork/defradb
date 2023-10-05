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
		if prop == request.Data { // parse data
			raw := argument.Value.(*ast.StringValue)
			if raw.Value == "" {
				return nil, ErrEmptyDataPayload
			}
			mut.Data = raw.Value
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
		} else if prop == request.DocID {
			raw := argument.Value.(*ast.StringValue)
			mut.IDs = immutable.Some([]string{raw.Value})
		} else if prop == request.DocIDs {
			raw := argument.Value.(*ast.ListValue)
			ids := make([]string, len(raw.Values))
			for i, val := range raw.Values {
				id, ok := val.(*ast.StringValue)
				if !ok {
					return nil, client.NewErrUnexpectedType[*ast.StringValue]("ids argument", val)
				}
				ids[i] = id.Value
			}
			mut.IDs = immutable.Some(ids)
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

	mut.Fields, err = parseSelectFields(schema, request.ObjectSelection, fieldObject, field.SelectionSet)
	return mut, err
}
