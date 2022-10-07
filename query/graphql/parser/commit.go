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

	"github.com/graphql-go/graphql/language/ast"

	"github.com/sourcenetwork/defradb/client"
	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"
)

var (
	_ Selection = (*CommitSelect)(nil)
)

type CommitSelect struct {
	Alias string
	Name  string

	DocKey    string
	FieldName client.Option[string]
	Cid       string

	Limit   *parserTypes.Limit
	OrderBy *parserTypes.OrderBy

	Fields []Selection
}

func (c CommitSelect) GetRoot() parserTypes.SelectionType {
	return parserTypes.CommitSelection
}

func (c CommitSelect) ToSelect() *Select {
	return &Select{
		Alias:   c.Alias,
		Limit:   c.Limit,
		OrderBy: c.OrderBy,
		Fields:  c.Fields,
		Root:    parserTypes.CommitSelection,
	}
}

func parseCommitSelect(field *ast.Field) (*CommitSelect, error) {
	commit := &CommitSelect{
		Name:  field.Name.Value,
		Alias: getFieldAlias(field),
	}

	for _, argument := range field.Arguments {
		prop := argument.Name.Value
		if prop == parserTypes.DocKey {
			raw := argument.Value.(*ast.StringValue)
			commit.DocKey = raw.Value
		} else if prop == parserTypes.Cid {
			raw := argument.Value.(*ast.StringValue)
			commit.Cid = raw.Value
		} else if prop == parserTypes.Field {
			raw := argument.Value.(*ast.StringValue)
			commit.FieldName = client.Some(raw.Value)
		} else if prop == parserTypes.OrderClause {
			obj := argument.Value.(*ast.ObjectValue)
			cond, err := ParseConditionsInOrder(obj)
			if err != nil {
				return nil, err
			}
			commit.OrderBy = &parserTypes.OrderBy{
				Conditions: cond,
				Statement:  obj,
			}
		} else if prop == parserTypes.LimitClause {
			val := argument.Value.(*ast.IntValue)
			limit, err := strconv.ParseInt(val.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			if commit.Limit == nil {
				commit.Limit = &parserTypes.Limit{}
			}
			commit.Limit.Limit = limit
		} else if prop == parserTypes.OffsetClause {
			val := argument.Value.(*ast.IntValue)
			offset, err := strconv.ParseInt(val.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			if commit.Limit == nil {
				commit.Limit = &parserTypes.Limit{}
			}
			commit.Limit.Offset = offset
		}
	}

	// latestCommits is just syntax sugar around a commits query
	if commit.Name == parserTypes.LatestCommitsQueryName {
		// Limit is not exposed as an input parameter for latestCommits,
		// so we can blindly set it here without worrying about existing
		// values
		commit.Limit = &parserTypes.Limit{
			Limit: 1,
		}

		// OrderBy is not exposed as an input parameter for latestCommits,
		// so we can blindly set it here without worrying about existing
		// values
		commit.OrderBy = &parserTypes.OrderBy{
			Conditions: []parserTypes.OrderCondition{
				{
					Field:     parserTypes.HeightFieldName,
					Direction: parserTypes.DESC,
				},
			},
		}
	}

	// no sub fields (unlikely)
	if field.SelectionSet == nil {
		return commit, nil
	}

	var err error
	commit.Fields, err = parseSelectFields(commit.GetRoot(), field.SelectionSet)

	return commit, err
}
