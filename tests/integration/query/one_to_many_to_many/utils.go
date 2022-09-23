// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many_to_many

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var gqlSchemaOneToManyToMany = (`

	type Author {
		name: String
		age: Int
		verified: Boolean
		book: [Book]
	}

	type Book {
		name: String
		rating: Float
		author: Author
        publisher: [Publisher]
	}

    type Publisher {
        name: String
        address: String
        yearOpened: Int
        book: Book
    }

`)

func executeTestCase(
	t *testing.T,
	test testUtils.QueryTestCase,
) {
	testUtils.ExecuteQueryTestCase(
		t,
		gqlSchemaOneToManyToMany,
		[]string{
			"Author",
			"Book",
			"Publisher",
		},
		test,
	)
}
