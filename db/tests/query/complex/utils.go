// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package complex

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/db/tests"
)

var bookAuthorPublisherGQLSchema = (`
	type book {
		name: String
		rating: Float
		author: author
		publisher: publisher
	}

	type author {
		name: String
		age: Int
		verified: Boolean
		wrote: book @primary
	}

	type publisher {
		name: String
		address: String
		published: book
	}
`)

func executeTestCase(t *testing.T, test testUtils.QueryTestCase) {
	testUtils.ExecuteQueryTestCase(
		t,
		bookAuthorPublisherGQLSchema,
		[]string{"book", "author", "publisher"},
		test,
	)
}
