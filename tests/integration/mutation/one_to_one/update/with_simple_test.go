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
							update_Author(data: "{\"name\": \"John Grisham\",\"published_id\": \"bae-fd541c25-229e-5280-b44b-e5c2af3e374d\"}") {
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

func TestMutationUpdateOneToOne(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one update mutation",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.Request{
				Request: `
				mutation {
					update_Author(data: "{\"name\": \"John Grisham\",\"published_id\": \"bae-3d236f89-6a31-5add-a36a-27971a2eac76\"}") {
						name
					}
				}`,
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
			testUtils.Request{
				Request: `
					query {
						Author {
							name
							published {
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"name": "John Grisham",
						"published": map[string]any{
							"name": "Painted House",
						},
					},
				},
			},
		},
	}

	simpleTests.ExecuteTestCase(t, test)
}

func TestMutationUpdateOneToOneSecondarySide(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, from the secondary side",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.Request{
				Request: `
				mutation {
					update_Book(data: "{\"name\": \"Painted House\",\"author_id\": \"bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed\"}") {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Painted House",
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
			testUtils.Request{
				Request: `
					query {
						Author {
							name
							published {
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"name": "John Grisham",
						"published": map[string]any{
							"name": "Painted House",
						},
					},
				},
			},
		},
	}
	simpleTests.ExecuteTestCase(t, test)
}
