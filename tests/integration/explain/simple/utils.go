// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

type dataMap = map[string]any

var bookAuthorGQLSchema = (`
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

`)

// TODO: This should be resolved in https://github.com/sourcenetwork/defradb/issues/953.
func executeTestCase(t *testing.T, test testUtils.RequestTestCase) {
	testUtils.ExecuteRequestTestCase(
		t,
		bookAuthorGQLSchema,
		[]string{"Article", "Book", "Author", "AuthorContact", "AontactAddress"},
		test,
	)
}

// TODO: This comment is removed in PR that resolves https://github.com/sourcenetwork/defradb/issues/953
//func executeExplainTestCase(t *testing.T, test explainUtils.ExplainRequestTestCase) {
//	explainUtils.ExecuteExplainRequestTestCase(
//		t,
//		bookAuthorGQLSchema,
//		[]string{"article", "book", "author", "authorContact", "contactAddress"},
//		test,
//	)
//}
