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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	simpleTests "github.com/sourcenetwork/defradb/tests/integration/mutation/simple"
)

func TestMutationCreateSimple(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple create mutation",
		Query: `mutation {
					create_user(data: "{\"name\": \"John\",\"age\": 27,\"points\": 42.1,\"verified\": true}") {
						_key
						name
						age
					}
				}`,
		Docs: map[int][]string{},
		Results: []map[string]interface{}{
			{
				"_key": "bae-0a24cf29-b2c2-5861-9d00-abd6250c475d",
				"age":  int64(27),
				"name": "John",
			},
		},
	}

	simpleTests.ExecuteTestCase(t, test)
}

func TestMutationCreateSimpleDoesNotCreateDocGivenDuplicate(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "Simple create mutation where document already exists.",
		Query: `mutation {
					create_user(data: "{\"name\": \"John\",\"age\": 27}") {
						_key
						name
						age
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"name": "John",
				"age": 27
			}`)},
		},
		ExpectedError: "A document with the given key already exists",
	}

	simpleTests.ExecuteTestCase(t, test)
}
