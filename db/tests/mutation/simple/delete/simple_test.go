// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package delete

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/db/tests"
	simpleTests "github.com/sourcenetwork/defradb/db/tests/mutation/simple"
)

func TestMutationDeleteDocumentUsingSingleKey(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "Simple delete mutation while no document exists.",
			Query: `mutation {
						delete_user(id: "bae-028383cc-d6ba-5df7-959f-2bdce3536a05") {
							_key
						}
					}`,
			Docs:          map[int][]string{},
			Results:       nil,
			ExpectedError: "No document for the given key exists",
		},

		{
			Description: "Simple delete mutation where one element exists.",
			Query: `mutation {
						delete_user(id: "bae-8ca944fd-260e-5a44-b88f-326d9faca810") {
							_key
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
						"name": "Shahzad",
						"age":  26,
						"points": 48.5,
						"verified": true
					}`)},
			},
			Results: []map[string]interface{}{
				{
					"_key": "bae-8ca944fd-260e-5a44-b88f-326d9faca810",
				},
			},
		},
	}

	for _, test := range tests {
		simpleTests.ExecuteTestCase(t, test)
	}
}
