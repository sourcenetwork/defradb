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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var schema = `
	type Users {
		name: String
		age: Int
		heightM: Float
		verified: Boolean
	}
`

func executeTestCase(t *testing.T, test testUtils.TestCase) {
	testUtils.ExecuteTestCase(
		t,
		testUtils.TestCase{
			Description: test.Description,
			Actions: append(
				[]any{
					testUtils.SchemaUpdate{
						Schema: schema,
					},
				},
				test.Actions...,
			),
		},
	)
}
