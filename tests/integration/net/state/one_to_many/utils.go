// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration/net/state"
)

var schema = (`
	type Author {
		Name: String
		Books: [Book]
	}
	type Book {
		Name: String
		Author: Author
	}
`)

func ExecuteTestCase(t *testing.T, test testUtils.P2PTestCase) {
	testUtils.ExecuteTestCase(t, schema, []string{"Author", "Book"}, test)
}
