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
	"github.com/sourcenetwork/defradb/core"
	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"
)

var (
	_ Selection = (*CommitSelect)(nil)
)

type CommitSelect struct {
	Field

	DocKey    client.Option[string]
	FieldName client.Option[string]
	Cid       client.Option[string]
	Depth     client.Option[uint64]

	Limit   client.Option[uint64]
	Offset  client.Option[uint64]
	OrderBy *parserTypes.OrderBy
	GroupBy *parserTypes.GroupBy

	Fields []Selection
}

func (c CommitSelect) GetRoot() parserTypes.SelectionType {
	return parserTypes.CommitSelection
}

func (c CommitSelect) ToSelect() *Select {
	return &Select{
		Field: Field{
			Name:  c.Name,
			Alias: c.Alias,
		},
		Limit:   c.Limit,
		Offset:  c.Offset,
		OrderBy: c.OrderBy,
		GroupBy: c.GroupBy,
		Fields:  c.Fields,
		Root:    parserTypes.CommitSelection,
	}
}

func parseCommitSelect(field *ast.Field) (*CommitSelect, error) {
	commit := &CommitSelect{
		Field: Field{
			Name:  field.Name.Value,
			Alias: getFieldAlias(field),
		},
	}

	for _, argument := range field.Arguments {
		prop := argument.Name.Value
		if prop == parserTypes.DocKey {
			raw := argument.Value.(*ast.StringValue)
			commit.DocKey = client.Some(raw.Value)
		} else if prop == parserTypes.Cid {
			raw := argument.Value.(*ast.StringValue)
			commit.Cid = client.Some(raw.Value)
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
			}
		} else if prop == parserTypes.LimitClause {
			val := argument.Value.(*ast.IntValue)
			limit, err := strconv.ParseUint(val.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			commit.Limit = client.Some(limit)
		} else if prop == parserTypes.OffsetClause {
			val := argument.Value.(*ast.IntValue)
			offset, err := strconv.ParseUint(val.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			commit.Offset = client.Some(offset)
		} else if prop == parserTypes.DepthClause {
			raw := argument.Value.(*ast.IntValue)
			depth, err := strconv.ParseUint(raw.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			commit.Depth = client.Some(depth)
		} else if prop == parserTypes.GroupByClause {
			obj := argument.Value.(*ast.ListValue)
			fields := []string{}
			for _, v := range obj.Values {
				fields = append(fields, v.GetValue().(string))
			}

			commit.GroupBy = &parserTypes.GroupBy{
				Fields: fields,
			}
		}
	}

	// latestCommits is just syntax sugar around a commits query
	if commit.Name == parserTypes.LatestCommitsQueryName {
		// Depth is not exposed as an input parameter for latestCommits,
		// so we can blindly set it here without worrying about existing
		// values
		commit.Depth = client.Some(uint64(1))

		if !commit.FieldName.HasValue() {
			// latest commits defaults to composite commits only at the moment
			commit.FieldName = client.Some(core.COMPOSITE_NAMESPACE)
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
