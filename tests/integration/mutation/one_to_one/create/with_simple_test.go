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
	simpleTests "github.com/sourcenetwork/defradb/tests/integration/mutation/one_to_one"
)

func TestMutationCreateOneToOneWrongSide(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One to one create mutation, from the wrong side",
		Request: `mutation {
					create_book(data: "{\"name\": \"Painted House\",\"author_id\": \"bae-fd541c25-229e-5280-b44b-e5c2af3e374d\"}") {
						_key
					}
				}`,
		ExpectedError: "The given field does not exist",
	}

	simpleTests.ExecuteTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationCreateOneToOneNoChild(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One to one create mutation, from the wrong side",
		Request: `mutation {
					create_author(data: "{\"name\": \"John Grisham\",\"published_id\": \"bae-fd541c25-229e-5280-b44b-e5c2af3e374d\"}") {
						name
					}
				}`,
		Results: []map[string]any{
			{
				"name": "John Grisham",
			},
		},
	}

	simpleTests.ExecuteTestCase(t, test)
}
