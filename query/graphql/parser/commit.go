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
	"github.com/sourcenetwork/defradb/errors"
	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"
)

type CommitType int

const (
	NoneCommitType = CommitType(iota)
	LatestCommits
	Commits
)

var (
	commitNameToType = map[string]CommitType{
		"latestCommits": LatestCommits,
		"commits":       Commits,
	}

	_ Selection = (*CommitSelect)(nil)
)

type CommitSelect struct {
	Alias string
	Name  string

	Type      CommitType
	DocKey    string
	FieldName client.Option[string]
	Cid       string
	Depth     client.Option[uint64]

	Limit   *parserTypes.Limit
	OrderBy *parserTypes.OrderBy
	GroupBy *parserTypes.GroupBy

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
		GroupBy: c.GroupBy,
		Fields:  c.Fields,
		Root:    parserTypes.CommitSelection,
	}
}

func parseCommitSelect(field *ast.Field) (*CommitSelect, error) {
	commit := &CommitSelect{
		Name:  field.Name.Value,
		Alias: getFieldAlias(field),
	}

	var ok bool
	commit.Type, ok = commitNameToType[commit.Name]
	if !ok {
		return nil, errors.New("Unknown Database query")
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
		} else if prop == parserTypes.DepthClause {
			raw := argument.Value.(*ast.IntValue)
			depth, err := strconv.ParseUint(raw.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			commit.Depth = client.Some(depth)
		} else if prop == parserTypes.GroupByClause {
			obj := argument.Value.(*ast.ListValue)
			fields := make([]string, 0)
			for _, v := range obj.Values {
				fields = append(fields, v.GetValue().(string))
			}

			commit.GroupBy = &parserTypes.GroupBy{
				Fields: fields,
			}
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
