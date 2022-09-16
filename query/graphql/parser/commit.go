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
	"github.com/graphql-go/graphql/language/ast"
	"github.com/sourcenetwork/defradb/errors"
	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"
)

type CommitType int

const (
	NoneCommitType = CommitType(iota)
	LatestCommits
	AllCommits
	OneCommit
)

var (
	commitNameToType = map[string]CommitType{
		"latestCommits": LatestCommits,
		"allCommits":    AllCommits,
		"commit":        OneCommit,
	}

	_ Selection = (*CommitSelect)(nil)
)

type CommitSelect struct {
	Alias string
	Name  string

	Type      CommitType
	DocKey    string
	FieldName string
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
			commit.FieldName = raw.Value
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
