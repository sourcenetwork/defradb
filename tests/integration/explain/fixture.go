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

package test_explain

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var SchemaForExplainTests = &action.AddCollection{
	SDL: (`
		type Article {
			name: String
			author: Author
			pages: Int
		}

		type Book {
			name: String
			author: Author
			rating: Float
			pages: Int
			chapterPages: [Int!]
		}

		type Author {
			name: String
			age: Int
			verified: Boolean
			books: [Book]
			articles: [Article]
			contact: AuthorContact @primary
		}

		type AuthorContact {
			cell: String
			email: String
			author: Author
			address: ContactAddress @primary
		}

		type ContactAddress {
			city: String
			country: String
			contact: AuthorContact
		}
	`),
}

func ExecuteTestCase(t *testing.T, test testUtils.TestCase) {
	testUtils.ExecuteTestCase(
		t,
		test,
	)
}
