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

var myUserSchema = (`
	type My_User {
		name: String
	}
`)

func executeTestCase(t *testing.T, test testUtils.RequestTestCase) {
	testUtils.ExecuteRequestTestCase(t, myUserSchema, []string{"My_User"}, test)
}

func TestMutationUpdateUnderscoredSchema(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple update of schema with underscored name",
		Request: `mutation {
			update_My_User(data: "{\"name\": \"Fred\"}") {
				_key
				name
			}
		}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John"
				}`,
			},
		},
		Results: []map[string]any{
			{
				"_key": "bae-decf6467-4c7c-50d7-b09d-0a7097ef6bad",
				"name": "Fred",
			},
		},
	}

	executeTestCase(t, test)
}
