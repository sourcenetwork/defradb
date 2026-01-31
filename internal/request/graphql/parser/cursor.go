// Copyright 2024 Democratized Data Foundation
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
	gql "github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"

	"github.com/sourcenetwork/defradb/client/request"
)

// parseCursorSelect parses a _cursor query block into a CursorSelect.
// Exactly one collection query is allowed inside _cursor, plus optionally _pageInfo.
func parseCursorSelect(
	exe *gql.ExecutionContext,
	parent *gql.Object,
	field *ast.Field,
) (*request.CursorSelect, error) {
	cursor := &request.CursorSelect{
		Field: request.Field{
			Name:  field.Name.Value,
			Alias: getFieldAlias(field),
		},
	}

	if field.SelectionSet == nil {
		return cursor, nil
	}

	fieldDef := gql.GetFieldDef(exe.Schema, parent, field.Name.Value)
	if fieldDef == nil {
		return cursor, nil
	}

	cursorType, err := typeFromFieldDef(fieldDef)
	if err != nil {
		return nil, err
	}

	collectedFields := gql.CollectFields(gql.CollectFieldsParams{
		ExeContext:   exe,
		RuntimeType:  cursorType,
		SelectionSet: field.SelectionSet,
		Fields:       map[string][]*ast.Field{},
	})

	var collectionCount int
	for _, fields := range collectedFields {
		for _, innerField := range fields {
			fieldName := innerField.Name.Value

			if fieldName == request.PageInfoFieldName {
				continue
			}

			collectionCount++
			if collectionCount > 1 {
				return nil, request.ErrMultipleQueriesInCursor
			}

			innerFieldDef := gql.GetFieldDef(exe.Schema, cursorType, fieldName)
			if innerFieldDef == nil {
				continue
			}

			parsed, err := parseSelect(exe, cursorType, innerField)
			if err != nil {
				return nil, err
			}

			cursor.Select = parsed
		}
	}

	return cursor, nil
}
