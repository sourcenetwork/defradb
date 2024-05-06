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
	"strconv"

	gql "github.com/sourcenetwork/graphql-go"
	"github.com/sourcenetwork/graphql-go/language/ast"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/core"
)

func parseCommitSelect(schema gql.Schema, parent *gql.Object, field *ast.Field) (*request.CommitSelect, error) {
	commit := &request.CommitSelect{
		Field: request.Field{
			Name:  field.Name.Value,
			Alias: getFieldAlias(field),
		},
	}

	for _, argument := range field.Arguments {
		prop := argument.Name.Value
		if prop == request.DocIDArgName {
			raw := argument.Value.(*ast.StringValue)
			commit.DocID = immutable.Some(raw.Value)
		} else if prop == request.Cid {
			raw := argument.Value.(*ast.StringValue)
			commit.CID = immutable.Some(raw.Value)
		} else if prop == request.FieldIDName {
			raw := argument.Value.(*ast.StringValue)
			commit.FieldID = immutable.Some(raw.Value)
		} else if prop == request.OrderClause {
			obj := argument.Value.(*ast.ObjectValue)
			cond, err := ParseConditionsInOrder(obj)
			if err != nil {
				return nil, err
			}
			commit.OrderBy = immutable.Some(
				request.OrderBy{
					Conditions: cond,
				},
			)
		} else if prop == request.LimitClause {
			val := argument.Value.(*ast.IntValue)
			limit, err := strconv.ParseUint(val.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			commit.Limit = immutable.Some(limit)
		} else if prop == request.OffsetClause {
			val := argument.Value.(*ast.IntValue)
			offset, err := strconv.ParseUint(val.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			commit.Offset = immutable.Some(offset)
		} else if prop == request.DepthClause {
			raw := argument.Value.(*ast.IntValue)
			depth, err := strconv.ParseUint(raw.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			commit.Depth = immutable.Some(depth)
		} else if prop == request.GroupByClause {
			obj := argument.Value.(*ast.ListValue)
			fields := []string{}
			for _, v := range obj.Values {
				fields = append(fields, v.GetValue().(string))
			}

			commit.GroupBy = immutable.Some(
				request.GroupBy{
					Fields: fields,
				},
			)
		}
	}

	// latestCommits is just syntax sugar around a commits operation.
	if commit.Name == request.LatestCommitsName {
		// Depth is not exposed as an input parameter for latestCommits,
		// so we can blindly set it here without worrying about existing
		// values
		commit.Depth = immutable.Some(uint64(1))

		if !commit.FieldID.HasValue() {
			// latest commits defaults to composite commits only at the moment
			commit.FieldID = immutable.Some(core.COMPOSITE_NAMESPACE)
		}
	}

	// no sub fields (unlikely)
	if field.SelectionSet == nil {
		return commit, nil
	}

	fieldDef := gql.GetFieldDef(schema, parent, field.Name.Value)

	fieldObject, err := typeFromFieldDef(fieldDef)
	if err != nil {
		return nil, err
	}

	commit.Fields, err = parseSelectFields(schema, fieldObject, field.SelectionSet)

	return commit, err
}
