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
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
)

// parseMutationOperationDefinition parses the individual GraphQL
// 'mutation' operations, which there may be multiple of.
func parseMutationOperationDefinition(
	exe *executionContext,
	collectedFields [][]*field,
) (*request.OperationDefinition, error) {
	var selections []request.Selection
	for _, fields := range collectedFields {
		for _, node := range fields {
			mut, err := parseMutation(exe, node)
			if err != nil {
				return nil, err
			}
			selections = append(selections, mut)
		}
	}
	return &request.OperationDefinition{
		Selections: selections,
	}, nil
}

// @todo: Create separate mutation parse functions
// for generated object mutations, and general
// API mutations.

// parseMutation parses a typed mutation field
// which includes sub fields, and may include
// filters, IDs, payloads, etc.
func parseMutation(exe *executionContext, field *field) (*request.ObjectMutation, error) {
	mut := &request.ObjectMutation{
		Field: request.Field{
			Name:  field.name,
			Alias: field.alias,
		},
	}

	// parse the mutation type
	// mutation names are either generated from a type
	// which means they are in the form name_type, where
	// the name is the object mutation name (ie: add, update, delete)
	// or its an general API mutation, which is in the form
	// name (camelCase).
	// This means we can split on the "_" character, and always
	// get back one of our defined types.
	mutNameParts := strings.Split(mut.Name, "_")
	typeStr := mutNameParts[0]

	if len(mutNameParts) > 1 { // only generated object mutations
		// reconstruct the name.
		// if the schema/collection name is eg: my_book
		// then the mutation name would be add_my_book
		// so we need to recreate the string my_book, which
		// has been split by "_", so we just join by "_"
		mut.Collection = strings.Join(mutNameParts[1:], "_")
	}

	switch typeStr {
	case "add":
		mut.Type = request.AddObjects
		if err := parseAddMutationArgs(mut, field.arguments); err != nil {
			return nil, err
		}

	case "update":
		mut.Type = request.UpdateObjects
		parseUpdateMutationArgs(mut, field.arguments)

	case "delete":
		mut.Type = request.DeleteObjects
		parseDeleteMutationArgs(mut, field.arguments)

	case "upsert":
		mut.Type = request.UpsertObjects
		parseUpsertMutationArgs(mut, field.arguments)

	case "truncate":
		mut.Type = request.TruncateObjects
		if err := parseTruncateMutationArgs(mut, field.arguments); err != nil {
			return nil, err
		}

	default:
		return nil, ErrUnknownMutationName
	}
	resolveUTCNowMutation(mut, exe.now)

	// if theres no field selections, just return
	if len(field.selectionSet) == 0 {
		return mut, nil
	}

	fields, err := parseSelectFields(exe, field.selectionSet)
	if err != nil {
		return nil, err
	}
	mut.Fields = fields

	return mut, err
}

func resolveUTCNowMutation(mut *request.ObjectMutation, now time.Time) {
	for _, input := range mut.AddInput {
		resolveUTCNowMap(input, now)
	}
	resolveUTCNowMap(mut.UpdateInput, now)
}

func resolveUTCNowMap(value map[string]any, now time.Time) {
	for key, item := range value {
		switch item := item.(type) {
		case utcNowValue:
			value[key] = now
		case map[string]any:
			resolveUTCNowMap(item, now)
		case []any:
			for index, element := range item {
				if _, ok := element.(utcNowValue); ok {
					item[index] = now
				} else if nested, ok := element.(map[string]any); ok {
					resolveUTCNowMap(nested, now)
				}
			}
		}
	}
}

func parseTruncateMutationArgs(mut *request.ObjectMutation, args map[string]any) error {
	for name, value := range args {
		switch name {
		case request.FilterClause:
			if value == nil {
				return ErrTruncateFilterNull
			}
			v, ok := value.(map[string]any)
			if !ok {
				return ErrInvalidFilterConditions
			}
			mut.Filter = immutable.Some(request.Filter{Conditions: v})
		}
	}
	return nil
}

func parseAddMutationArgs(mut *request.ObjectMutation, args map[string]any) error {
	for name, value := range args {
		switch name {
		case request.Input:
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}
			inputs := make([]map[string]any, len(v))
			for i, v := range v {
				input, ok := v.(map[string]any)
				if !ok {
					return newErrExpectedNull(mut.Collection + "MutationInputArg!")
				}
				inputs[i] = input
			}
			mut.AddInput = inputs

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
	return nil
}

func parseDeleteMutationArgs(mut *request.ObjectMutation, args map[string]any) {
	for name, value := range args {
		switch name {
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

		case request.FilterClause:
			if v, ok := value.(map[string]any); ok {
				mut.Filter = immutable.Some(request.Filter{Conditions: v})
			}
		}
	}
}

func parseUpdateMutationArgs(mut *request.ObjectMutation, args map[string]any) {
	for name, value := range args {
		switch name {
		case request.Input:
			if v, ok := value.(map[string]any); ok {
				mut.UpdateInput = v
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

		case request.FilterClause:
			if v, ok := value.(map[string]any); ok {
				mut.Filter = immutable.Some(request.Filter{Conditions: v})
			}
		}
	}
}

func parseUpsertMutationArgs(mut *request.ObjectMutation, args map[string]any) {
	for name, value := range args {
		switch name {
		case request.AddInput:
			if v, ok := value.(map[string]any); ok {
				mut.AddInput = []map[string]any{v}
			}

		case request.UpdateInput:
			if v, ok := value.(map[string]any); ok {
				mut.UpdateInput = v
			}

		case request.FilterClause:
			if v, ok := value.(map[string]any); ok {
				mut.Filter = immutable.Some(request.Filter{Conditions: v})
			}
		}
	}
}
