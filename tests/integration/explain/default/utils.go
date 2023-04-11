// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_explain_default

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	explainUtils "github.com/sourcenetwork/defradb/tests/integration/explain"
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
		contact: authorContact
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

// TODO: This should be resolved in ISSUE#953 (github.com/sourcenetwork/defradb).
func executeTestCase(t *testing.T, test testUtils.RequestTestCase) {
	testUtils.ExecuteRequestTestCase(
		t,
		bookAuthorGQLSchema,
		[]string{"Article", "Book", "Author", "AuthorContact", "ContactAddress"},
		test,
	)
}

func runExplainTest(t *testing.T, test explainUtils.ExplainRequestTestCase) {
	explainUtils.ExecuteExplainRequestTestCase(
		t,
		bookAuthorGQLSchema,
		[]string{"Article", "Book", "Author", "AuthorContact", "ContactAddress"},
		test,
	)
}

var basicPattern = dataMap{
	"explain": dataMap{
		"selectTopNode": dataMap{
			"selectNode": dataMap{
				"scanNode": dataMap{},
			},
		},
	},
}
