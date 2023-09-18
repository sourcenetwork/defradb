// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package special

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationInvalidMutation(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple invalid mutation",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
					dostuff_User(data: "") {
						_key
					}
				}`,
				ExpectedError: "Cannot query field \"dostuff_User\" on type \"Mutation\".",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
