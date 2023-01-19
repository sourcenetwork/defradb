// Copyright 2023 Democratized Data Foundation
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

	testUtils "github.com/sourcenetwork/defradb/tests/integration/net/state"
)

var userCollectionGQLSchema = (`
	type Users {
		Name: String
		Email: String
		Age: Int
		Height: Float
		Verified: Boolean
	}
`)

func ExecuteTestCase(t *testing.T, test testUtils.P2PTestCase) {
	testUtils.ExecuteTestCase(t, userCollectionGQLSchema, []string{"Users"}, test)
}
