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

package one_to_many_to_one

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestOneToManyToOneDeepOrderBySubTypeOfBothDescAndAsc(t *testing.T) {
	test := testUtils.TestCase{
		// TODO: https://github.com/sourcenetwork/defradb/issues/4353
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			addDocsWith6BooksAnd5Publishers(),
			&action.Request{
				Request: `query {
					Author {
						name
						NewestPublishersBook: book(order: {publisher: {yearOpened: DESC}}, limit: 1) {
							name
						}
						OldestPublishersBook: book(order: {publisher: {yearOpened: ASC}}, limit: 1) {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "Cornelia Funke",
							"NewestPublishersBook": []map[string]any{
								{
									"name": "The Rooster Bar",
								},
							},
							"OldestPublishersBook": []map[string]any{
								{
									"name": "The Rooster Bar",
								},
							},
						},
						{
							"name": "John Grisham",
							"NewestPublishersBook": []map[string]any{
								{
									"name": "Theif Lord",
								},
							},
							"OldestPublishersBook": []map[string]any{
								{
									"name": "The Associate", // oldest because has no Publisher.
								},
							},
						},
						{
							"name":                 "Not a Writer",
							"NewestPublishersBook": []map[string]any{},
							"OldestPublishersBook": []map[string]any{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
