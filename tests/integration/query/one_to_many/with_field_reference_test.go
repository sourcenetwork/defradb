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

// A bare, unquoted identifier in a comparison position names another field of the same
// document.  These tests cover the aggregate and relation cases.

var authorBookCountGQLSchema = `
	type Book {
		name: String
		author: Author
	}

	type Author {
		name: String
		expectedBooks: Int
		published: [Book]
	}
`

func authorBookCountActions() []any {
	return []any{
		&action.AddCollection{SDL: authorBookCountGQLSchema},
		&action.AddDoc{
			CollectionID: 1,
			// {{.DocID1_0}} - has fewer books than expected
			Doc: `{
				"name": "John Grisham",
				"expectedBooks": 3
			}`,
		},
		&action.AddDoc{
			CollectionID: 1,
			// {{.DocID1_1}} - has as many books as expected
			Doc: `{
				"name": "Cornelia Funke",
				"expectedBooks": 1
			}`,
		},
		&action.AddDoc{
			CollectionID: 0,
			Doc: `{
				"name": "Painted House",
				"_authorID": "{{.DocID1_0}}"
			}`,
		},
		&action.AddDoc{
			CollectionID: 0,
			Doc: `{
				"name": "Theif Lord",
				"_authorID": "{{.DocID1_1}}"
			}`,
		},
	}
}

// An aggregate is compared against a field by aliasing the aggregate and filtering on it
// through `_alias`.
func TestQueryOneToManyWithFieldReferenceInAliasedAggregateFilter(t *testing.T) {
	actions := append(authorBookCountActions(), &action.Request{
		Request: `query {
			Author(filter: {_alias: {bookCount: {_lt: expectedBooks}}}) {
				name
				bookCount: COUNT(published: {})
			}
		}`,
		Results: map[string]any{
			"Author": []map[string]any{
				{
					"name":      "John Grisham",
					"bookCount": 1,
				},
			},
		},
	})

	testUtils.ExecuteTestCase(t, testUtils.TestCase{Actions: actions})
}

// A reference inside a relation sub-filter would name a field of the related document, which
// is not supported.  It must be rejected rather than silently matching nothing.
func TestQueryOneToManyWithFieldReferenceInsideRelationFilter(t *testing.T) {
	actions := append(authorBookCountActions(), &action.Request{
		Request: `query {
			Book(filter: {author: {expectedBooks: {_lt: name}}}) {
				name
			}
		}`,
		ExpectedError: `field-to-field comparison is only supported between fields of the same document`,
	})

	testUtils.ExecuteTestCase(t, testUtils.TestCase{Actions: actions})
}
