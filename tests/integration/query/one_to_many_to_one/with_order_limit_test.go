// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestOneToManyToOneDeepOrderBySubTypeOfBothDescAndAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "1-N-1 deep orderby subtypes of both descending and ascending.",
		Actions: []any{
			gqlSchemaOneToManyToOne(),
			createDocsWith6BooksAnd5Publishers(),
			testUtils.Request{
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
							"name":                 "Not a Writer",
							"NewestPublishersBook": []map[string]any{},
							"OldestPublishersBook": []map[string]any{},
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
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
