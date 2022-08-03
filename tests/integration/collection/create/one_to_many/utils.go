// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package create

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration/collection"
)

var schema = (`
	type book {
		Name: String
		Rating: Float
		Author: author
	}

	type author {
		Name: String
		Age: Int
		Verified: Boolean
		Published: [book]
	}
`)

func executeTestCase(t *testing.T, test testUtils.TestCase) {
	testUtils.ExecuteQueryTestCase(t, schema, test)
}
