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
	exe *gql.ExecutionContext,
	def *ast.OperationDefinition,
) (*request.OperationDefinition, error) {
	qdef := &request.OperationDefinition{
		Selections: make([]request.Selection, len(def.SelectionSet.Selections)),
	}

	for i, selection := range def.SelectionSet.Selections {
		switch node := selection.(type) {
		case *ast.Field:
			mut, err := parseMutation(exe, exe.Schema.MutationType(), node)
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
func parseMutation(exe *gql.ExecutionContext, parent *gql.Object, field *ast.Field) (*request.ObjectMutation, error) {
	mut := &request.ObjectMutation{
		Field: request.Field{
			Name:  field.Name.Value,
			Alias: getFieldAlias(field),
		},
	}

	fieldDef := gql.GetFieldDef(exe.Schema, parent, mut.Name)
	arguments := gql.GetArgumentValues(fieldDef.Args, field.Arguments, exe.VariableValues)

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

	for _, argument := range field.Arguments {
		name := argument.Name.Value
		value := arguments[name]

		switch name {
		case request.Input:
			switch v := value.(type) {
			case []any:
				// input for create is a list
				inputs := make([]map[string]any, len(v))
				for i, v := range v {
					inputs[i] = v.(map[string]any)
				}
				mut.Input = inputs

			case map[string]any:
				// input for update is an object
				mut.Input = []map[string]any{v}
			}

		case request.FilterClause:
			if v, ok := value.(map[string]any); ok {
				mut.Filter = immutable.Some(request.Filter{Conditions: v})
			}

		case request.DocIDArgName:
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}
			docIDs := make([]string, len(v))
			for i, v := range v {
				docIDs[i] = v.(string)
			}
			mut.DocIDs = immutable.Some(docIDs)

		case request.EncryptDocArgName:
			if v, ok := value.(bool); ok {
				mut.Encrypt = v
			}

		case request.EncryptFieldsArgName:
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}
			fields := make([]string, len(v))
			for i, v := range v {
				fields[i] = v.(string)
			}
			mut.EncryptFields = fields
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

	mut.Fields, err = parseSelectFields(exe, fieldObject, field.SelectionSet)
	if err != nil {
		return nil, err
	}

	return mut, err
}
