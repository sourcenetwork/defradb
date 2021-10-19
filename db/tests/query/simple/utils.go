// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/db/tests"
)

var userCollectionGQLSchema = (`
	type users {
		Name: String
		Age: Int
		Verified: Boolean
	}
`)

func executeTestCase(t *testing.T, test testUtils.QueryTestCase) {
	testUtils.ExecuteQueryTestCase(t, userCollectionGQLSchema, []string{"users"}, test)
}
