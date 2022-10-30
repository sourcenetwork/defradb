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

	gql "github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
)

var (
	ErrEmptyDataPayload = errors.New("given data payload is empty")
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
func parseMutationOperationDefinition(schema gql.Schema, def *ast.OperationDefinition) (*request.OperationDefinition, error) {
	qdef := &request.OperationDefinition{
		Selections: make([]request.Selection, len(def.SelectionSet.Selections)),
	}

	qdef.IsExplain = parseExplainDirective(def.Directives)

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
func parseMutation(schema gql.Schema, parent *gql.Object, field *ast.Field) (*request.Mutation, error) {
	mut := &request.Mutation{
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
		return nil, errors.New("unknown mutation name")
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
				return nil, errors.New("couldn't get argument type for filter")
			}
			filter, err := NewFilter(obj, filterType)
			if err != nil {
				return nil, err
			}

			mut.Filter = filter
		} else if prop == request.Id {
			raw := argument.Value.(*ast.StringValue)
			mut.IDs = client.Some([]string{raw.Value})
		} else if prop == request.Ids {
			raw := argument.Value.(*ast.ListValue)
			ids := make([]string, len(raw.Values))
			for i, val := range raw.Values {
				id, ok := val.(*ast.StringValue)
				if !ok {
					return nil, errors.New("ids argument has a non string value")
				}
				ids[i] = id.Value
			}
			mut.IDs = client.Some(ids)
		}
	}

	// if theres no field selections, just return
	if field.SelectionSet == nil {
		return mut, nil
	}

	var fieldObject *gql.Object
	switch ftype := fieldDef.Type.(type) {
	case *gql.Object:
		fieldObject = ftype
	case *gql.List:
		fieldObject = ftype.OfType.(*gql.Object)
	default:
		return nil, errors.New("Couldn't get field object from definition")
	}
	var err error
	mut.Fields, err = parseSelectFields(schema, request.ObjectSelection, fieldObject, field.SelectionSet)
	return mut, err
}
