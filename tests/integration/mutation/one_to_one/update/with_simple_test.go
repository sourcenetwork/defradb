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
	simpleTests "github.com/sourcenetwork/defradb/tests/integration/mutation/one_to_one"
)

// This test documents incorrect behaviour. It should be possible to update book to point to author.
// https://github.com/sourcenetwork/defradb/issues/1214
func TestMutationUpdateOneToOneWrongSide(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, from the wrong side",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord"
				}`,
			},
			testUtils.Request{
				Request: `mutation {
							update_book(data: "{\"name\": \"Painted House\",\"author_id\": \"bae-fd541c25-229e-5280-b44b-e5c2af3e374d\"}") {
								_key
							}
						}`,
				ExpectedError: "The given field does not exist",
			},
		},
	}
	simpleTests.ExecuteTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationUpdateOneToOneNoChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, from the wrong side",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.Request{
				Request: `mutation {
							update_author(data: "{\"name\": \"John Grisham\",\"published_id\": \"bae-fd541c25-229e-5280-b44b-e5c2af3e374d\"}") {
								name
							}
						}`,
				Results: []map[string]any{
					{
						"name": "John Grisham",
					},
				},
			},
		},
	}
	simpleTests.ExecuteTestCase(t, test)
}
