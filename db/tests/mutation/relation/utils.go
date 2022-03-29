// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package relation_test

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/db/tests"
)

var bookSchema = (`
    type book {
        name: String
        rating: Float
        author: author @primary
    }

    type author {
        name: String
        age: Int
        verified: Boolean
        published: book
    }
`)

func ExecuteTestCase(t *testing.T, test testUtils.QueryTestCase) {
	testUtils.ExecuteQueryTestCase(t, bookSchema, []string{"book", "author"}, test)
}
