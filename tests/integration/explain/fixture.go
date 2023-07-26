// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var SchemaForExplainTests = testUtils.SchemaUpdate{
	Schema: (`
		type Article {
			name: String
			author: Author
			pages: Int
		}

		type Book {
			name: String
			author: Author
			pages: Int
			chapterPages: [Int!]
		}

		type Author {
			name: String
			age: Int
			verified: Boolean
			books: [Book]
			articles: [Article]
			contact: AuthorContact
		}

		type AuthorContact {
			cell: String
			email: String
			author: Author
			address: ContactAddress
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
