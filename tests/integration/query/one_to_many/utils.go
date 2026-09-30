// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package one_to_many

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var bookAuthorGQLSchema = (`
	type Book {
		name: String
		rating: Float
		author: Author
	}

	type Author {
		name: String
		age: Int
		verified: Boolean
		published: [Book]
	}
`)

func executeTestCase(t *testing.T, test testUtils.TestCase) {
	test.Actions = append(
		[]any{
			&action.AddCollection{
				SDL: bookAuthorGQLSchema,
			},
		},
		orderInitialDocs(test.Actions)...,
	)
	testUtils.ExecuteTestCase(t, test)
}

func orderInitialDocs(actions []any) []any {
	start := 0
	if len(actions) > 0 {
		if _, ok := actions[0].(*action.AddCollection); ok {
			start = 1
		}
	}

	firstNonAdd := start
	for firstNonAdd < len(actions) {
		if _, ok := actions[firstNonAdd].(*action.AddDoc); !ok {
			break
		}
		firstNonAdd++
	}
	if firstNonAdd == start {
		return actions
	}

	ordered := make([]any, 0, len(actions))
	ordered = append(ordered, actions[:start]...)
	for _, item := range actions[start:firstNonAdd] {
		if add, ok := item.(*action.AddDoc); ok && add.CollectionID == 1 {
			ordered = append(ordered, item)
		}
	}
	for _, item := range actions[start:firstNonAdd] {
		if add, ok := item.(*action.AddDoc); ok && add.CollectionID != 1 {
			ordered = append(ordered, item)
		}
	}
	return append(ordered, actions[firstNonAdd:]...)
}
