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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
)

// parseCursorSelect parses a _cursor query block into a CursorSelect.
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
				cursor.PageInfoSelect = parsePageInfoSelect(innerField)
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

			arguments := gql.GetArgumentValues(innerFieldDef.Args, innerField.Arguments, exe.VariableValues)
			for _, argument := range innerField.Arguments {
				name := argument.Name.Value
				value := arguments[name]

				switch name {
				case request.FirstClause:
					if v, ok := value.(int32); ok {
						if v < 0 {
							return nil, request.ErrFirstMustBeNonNegative
						}
						cursor.First = immutable.Some(uint64(v))
					}
				case request.AfterClause:
					if v, ok := value.(string); ok {
						cursor.After = immutable.Some(v)
					}
				case request.LastClause:
					if v, ok := value.(int32); ok {
						if v < 0 {
							return nil, request.ErrLastMustBeNonNegative
						}
						cursor.Last = immutable.Some(uint64(v))
					}
				case request.BeforeClause:
					if v, ok := value.(string); ok {
						cursor.Before = immutable.Some(v)
					}
				}
			}
		}
	}

	return cursor, nil
}

// parsePageInfoSelect extracts which _pageInfo fields were selected.
func parsePageInfoSelect(field *ast.Field) *request.PageInfoSelect {
	if field.SelectionSet == nil {
		return nil
	}

	pageInfo := &request.PageInfoSelect{}
	for _, selection := range field.SelectionSet.Selections {
		if f, ok := selection.(*ast.Field); ok {
			switch f.Name.Value {
			case request.HasNextFieldName:
				pageInfo.HasNext = true
			case request.HasPrevFieldName:
				pageInfo.HasPrev = true
			case request.StartCursorFieldName:
				pageInfo.StartCursor = true
			case request.EndCursorFieldName:
				pageInfo.EndCursor = true
			}
		}
	}
	return pageInfo
}
