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
	"fmt"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	simpleTests "github.com/sourcenetwork/defradb/tests/integration/mutation/one_to_one"
)

// This test documents incorrect behaviour. It should be possible to create author then book,
// linking in the second create step (like in [TestMutationCreateOneToOne]).
// https://github.com/sourcenetwork/defradb/issues/1213
func TestMutationCreateOneToOneWrongSide(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, from the wrong side",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
							create_Book(data: "{\"name\": \"Painted House\",\"author_id\": \"bae-fd541c25-229e-5280-b44b-e5c2af3e374d\"}") {
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
func TestMutationCreateOneToOneNoChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, from the wrong side",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
							create_Author(data: "{\"name\": \"John Grisham\",\"published_id\": \"bae-fd541c25-229e-5280-b44b-e5c2af3e374d\"}") {
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

func TestMutationCreateOneToOne(t *testing.T) {
	bookKey := "bae-3d236f89-6a31-5add-a36a-27971a2eac76"

	test := testUtils.TestCase{
		Description: "One to one create mutation",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
						create_Book(data: "{\"name\": \"Painted House\"}") {
							_key
						}
					}`,
				Results: []map[string]any{
					{
						"_key": bookKey,
					},
				},
			},
			testUtils.Request{
				Request: fmt.Sprintf(
					`mutation {
						create_Author(data: "{\"name\": \"John Grisham\",\"published_id\": \"%s\"}") {
							name
						}
					}`,
					bookKey),
				Results: []map[string]any{
					{
						"name": "John Grisham",
					},
				},
			},
			testUtils.Request{
				Request: `
					query {
						Book {
							name
							author {
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"name": "Painted House",
						"author": map[string]any{
							"name": "John Grisham",
						},
					},
				},
			},
		},
	}

	simpleTests.ExecuteTestCase(t, test)
}
