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
	"errors"

	"github.com/graphql-go/graphql/language/ast"

	parserTypes "github.com/sourcenetwork/defradb/request/graphql/parser/types"
)

type OperationDefinition struct {
	Name       string
	Selections []Selection
	Statement  *ast.OperationDefinition
	IsExplain  bool
}

func (q OperationDefinition) GetStatement() ast.Node {
	return q.Statement
}

type Selection interface {
	Statement
	GetName() string
	GetAlias() string
	GetSelections() []Selection
	GetRoot() parserTypes.SelectionType
}

type Request struct {
	Queries   []*OperationDefinition
	Mutations []*OperationDefinition
	Statement *ast.Document
}

func (q Request) GetStatement() ast.Node {
	return q.Statement
}

// ParseRequest parses a root ast.Document, and returns a formatted Query object.
// Requires a non-nil doc, will error if given a nil doc
func ParseRequest(doc *ast.Document) (*Request, error) {
	if doc == nil {
		return nil, errors.New("ParseQuery requires a non nil ast.Document")
	}
	q := &Request{
		Statement: doc,
		Queries:   make([]*OperationDefinition, 0),
		Mutations: make([]*OperationDefinition, 0),
	}

	for _, def := range q.Statement.Definitions {
		switch node := def.(type) {
		case *ast.OperationDefinition:
			if node.Operation == "query" {
				// parse query operation definition.
				qdef, err := parseQueryOperationDefinition(node)
				if err != nil {
					return nil, err
				}
				q.Queries = append(q.Queries, qdef)
			} else if node.Operation == "mutation" {
				// parse mutation operation definition.
				mdef, err := parseMutationOperationDefinition(node)
				if err != nil {
					return nil, err
				}
				q.Mutations = append(q.Mutations, mdef)
			} else {
				return nil, errors.New("Unkown graphql operation type")
			}
		}
	}

	return q, nil
}

// parseExplainDirective returns true if we parsed / detected the explain directive label
// in this ast, and false otherwise.
func parseExplainDirective(directives []*ast.Directive) bool {

	// Iterate through all directives and ensure that the directive is at there.
	// - Note: the location we don't need to worry about as the schema takes care of it, as when
	//         request is made there will be a syntax error for directive usage at the wrong location,
	//         unless we add another directive named `@explain` at another location (which we should not).
	for _, directive := range directives {
		// The arguments pased to the directive are at `directive.Arguments`.
		if directive.Name.Value == parserTypes.ExplainLabel {
			return true
		}
	}

	return false
}
