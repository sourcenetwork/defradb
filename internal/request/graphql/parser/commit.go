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
	gql "github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
)

func parseCommitSelect(
	exe *gql.ExecutionContext,
	parent *gql.Object,
	field *ast.Field,
) (*request.CommitSelect, error) {
	commit := &request.CommitSelect{
		Field: request.Field{
			Name:  field.Name.Value,
			Alias: getFieldAlias(field),
		},
	}

	fieldDef := gql.GetFieldDef(exe.Schema, parent, field.Name.Value)
	arguments := gql.GetArgumentValues(fieldDef.Args, field.Arguments, exe.VariableValues)

	for _, argument := range field.Arguments {
		name := argument.Name.Value
		value := arguments[name]

		switch name {
		case request.DocIDArgName:
			if v, ok := value.(string); ok {
				commit.DocID = immutable.Some(v)
			}

		case request.CidFieldName:
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}

			if len(v) > 1 {
				// todo - This limitiation is temporary and should be removed in
				// https://github.com/sourcenetwork/defradb/issues/4303
				return nil, ErrMultipleCidsNotSupported
			}

			cids := make([]string, len(v))
			for i, value := range v {
				cids[i] = value.(string)
			}
			commit.CIDs = immutable.Some(cids)

		case request.OrderClause:
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}
			conditions, err := parseOrderConditionList(v)
			if err != nil {
				return nil, err
			}
			commit.OrderBy = immutable.Some(request.OrderBy{
				Conditions: conditions,
			})

		case request.LimitClause:
			if v, ok := value.(int32); ok {
				commit.Limit = immutable.Some(uint64(v))
			}

		case request.OffsetClause:
			if v, ok := value.(int32); ok {
				commit.Offset = immutable.Some(uint64(v))
			}

		case request.DepthClause:
			if v, ok := value.(int32); ok {
				commit.Depth = immutable.Some(uint64(v))
			}

		case request.GroupByClause:
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}
			fields := make([]string, len(v))
			for i, c := range v {
				fields[i] = c.(string)
			}
			commit.GroupBy = immutable.Some(request.GroupBy{
				Fields: fields,
			})

		case request.FilterClause:
			if v, ok := value.(map[string]any); ok {
				commit.Filter = immutable.Some(request.Filter{Conditions: v})
			}
		}
	}

	// no sub fields (unlikely)
	if field.SelectionSet == nil {
		return commit, nil
	}

	fieldObject, err := typeFromFieldDef(fieldDef)
	if err != nil {
		return nil, err
	}

	commit.Fields, err = parseSelectFields(exe, fieldObject, field.SelectionSet)
	if err != nil {
		return nil, err
	}

	return commit, err
}
