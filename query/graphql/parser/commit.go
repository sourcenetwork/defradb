// Copyright 2020 Source Inc.
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
	"errors"

	"github.com/graphql-go/graphql/language/ast"
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

	Limit   *Limit
	OrderBy *OrderBy

	Fields []Selection

	Statement *ast.Field
}

func (c CommitSelect) GetRoot() SelectionType {
	return CommitSelection
}

func (c CommitSelect) GetStatement() ast.Node {
	return c.Statement
}

func (c CommitSelect) GetName() string {
	return c.Name
}

func (c CommitSelect) GetAlias() string {
	return c.Alias
}

func (c CommitSelect) GetSelections() []Selection {
	return c.Fields
}

func (c CommitSelect) ToSelect() *Select {
	return &Select{
		Name:      c.Name,
		Alias:     c.Alias,
		Limit:     c.Limit,
		OrderBy:   c.OrderBy,
		Statement: c.Statement,
		Fields:    c.Fields,
		Root:      CommitSelection,
	}
}

func parseCommitSelect(field *ast.Field) (*CommitSelect, error) {
	commit := &CommitSelect{
		Statement: field,
	}
	commit.Name = field.Name.Value
	if field.Alias != nil {
		commit.Alias = field.Alias.Value
	}

	var ok bool
	commit.Type, ok = commitNameToType[commit.Name]
	if !ok {
		return nil, errors.New("Unknown Database query")
	}

	for _, argument := range field.Arguments {
		prop := argument.Name.Value
		if prop == "dockey" {
			raw := argument.Value.(*ast.StringValue)
			commit.DocKey = raw.Value
		} else if prop == "cid" {
			raw := argument.Value.(*ast.StringValue)
			commit.Cid = raw.Value
		} else if prop == "field" {
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
