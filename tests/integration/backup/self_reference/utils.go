// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package backup

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var schemas = (`
	type User {
		name: String
		age: Int
		boss: User @primary
	}
`)

func executeTestCase(t *testing.T, test testUtils.TestCase) {
	testUtils.ExecuteTestCase(
		t,
		[]string{"User"},
		testUtils.TestCase{
			Description: test.Description,
			Actions: append(
				[]any{
					testUtils.SchemaUpdate{
						Schema: schemas,
					},
				},
				test.Actions...,
			),
		},
	)
}
