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

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionDuplicateRelationNameWithoutRelationDirective(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						title: String
						author: Person
						reviewer: Person
					}

					type Person {
						name: String
						authoredBooks: [Book]
						reviewedBooks: [Book]
					}
				`,
				ExpectedError: "relation name is not unique within collection. Field: author, RelationName: book_person",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionDuplicateRelationNameWithRelationDirective(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						title: String
						author: Person @relation(name: "book_author")
						reviewer: Person @relation(name: "book_reviewer")
					}

					type Person {
						name: String
						authoredBooks: [Book] @relation(name: "book_author")
						reviewedBooks: [Book] @relation(name: "book_reviewer")
					}
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionSelfRelationPrimarySecondaryPairWithSameRelationName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Person {
						name: String
						friend: Person
						friends: [Person]
					}
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
