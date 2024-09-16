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
	"github.com/sourcenetwork/defradb/internal/core"
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
		prop := argument.Name.Value
		if prop == request.DocIDArgName {
			commit.DocID = immutable.Some(arguments[prop].(string))
		} else if prop == request.Cid {
			commit.CID = immutable.Some(arguments[prop].(string))
		} else if prop == request.FieldIDName {
			commit.FieldID = immutable.Some(arguments[prop].(string))
		} else if prop == request.OrderClause {
			conditions, err := ParseConditionsInOrder(argument.Value.(*ast.ObjectValue), arguments[prop].(map[string]any))
			if err != nil {
				return nil, err
			}
			commit.OrderBy = immutable.Some(request.OrderBy{
				Conditions: conditions,
			})
		} else if prop == request.LimitClause {
			commit.Limit = immutable.Some(uint64(arguments[prop].(int32)))
		} else if prop == request.OffsetClause {
			commit.Offset = immutable.Some(uint64(arguments[prop].(int32)))
		} else if prop == request.DepthClause {
			commit.Depth = immutable.Some(uint64(arguments[prop].(int32)))
		} else if prop == request.GroupByClause {
			fields := []string{}
			for _, v := range arguments[prop].([]any) {
				fields = append(fields, v.(string))
			}
			commit.GroupBy = immutable.Some(request.GroupBy{
				Fields: fields,
			})
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

	fieldObject, err := typeFromFieldDef(fieldDef)
	if err != nil {
		return nil, err
	}

	commit.Fields, err = parseSelectFields(exe, fieldObject, field.SelectionSet)

	return commit, err
}
