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
	type article {
		name: String
		author: author
		pages: Int
	}

	type book {
		name: String
		author: author
		pages: Int
		chapterPages: [Int!]
	}

	type author {
		name: String
		age: Int
		verified: Boolean
		books: [book]
		articles: [article]
		contact: authorContact
	}

	type authorContact {
		cell: String
		email: String
		author: author
		address: contactAddress
	}

	type contactAddress {
		city: String
		country: String
		contact: authorContact
	}

`)

// TODO: This should be resolved in ISSUE#953 (github.com/sourcenetwork/defradb).
func executeTestCase(t *testing.T, test testUtils.RequestTestCase) {
	testUtils.ExecuteQueryTestCase(
		t,
		bookAuthorGQLSchema,
		[]string{"article", "book", "author", "authorContact", "contactAddress"},
		test,
	)
}

func executeExplainTestCase(t *testing.T, test explainUtils.ExplainRequestTestCase) {
	explainUtils.ExecuteExplainRequestTestCase(
		t,
		bookAuthorGQLSchema,
		[]string{"article", "book", "author", "authorContact", "contactAddress"},
		test,
	)
}
