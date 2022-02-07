// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package one_to_two_many

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/db/tests"
)

var bookAuthorGQLSchema = (`
	type book {
		name: String
		rating: Float
		price: price
		author: author @relation(name: "written_books")
		reviewedBy: author @relation(name: "reviewed_books")
	}

	type author {
		name: String
		age: Int
		verified: Boolean
		written: [book] @relation(name: "written_books")
		reviewed: [book] @relation(name: "reviewed_books")
	}

	type price {
		currency: String
		value: Float
		books: [book]
	}
`)

func executeTestCase(t *testing.T, test testUtils.QueryTestCase) {
	testUtils.ExecuteQueryTestCase(t, bookAuthorGQLSchema, []string{"book", "author", "price"}, test)
}
