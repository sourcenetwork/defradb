// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package update

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration/collection"
)

var schema = `
	type Book {
		name: String
		rating: Float
		author: Author
	}

	type Author {
		name: String
		age: Int
		verified: Boolean
		published: Book @primary
	}
`

func executeTestCase(t *testing.T, test testUtils.TestCase) {
	testUtils.ExecuteRequestTestCase(t, schema, test)
}
