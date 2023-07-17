// Copyright 2023 Democratized Data Foundation
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

func TestMutationCreateOneToOne_UseAliasWithInvalidField_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, alias relation, with an invalid field.",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
					create_Author(data: "{\"notName\": \"John Grisham\",\"published\": \"bae-fd541c25-229e-5280-b44b-e5c2af3e374d\"}") {
					name
				}
			}`,
				ExpectedError: "The given field does not exist. Name: notName",
			},
		},
	}
	simpleTests.ExecuteTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationCreateOneToOne_UseAliasWithNonExistingRelationPrimarySide_CreatedDoc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, alias relation, from the wrong side",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
					create_Author(data: "{\"name\": \"John Grisham\",\"published\": \"bae-fd541c25-229e-5280-b44b-e5c2af3e374d\"}") {
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

func TestMutationCreateOneToOne_UseAliasWithNonExistingRelationSecondarySide_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, alias relation, from the secondary side",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
					create_Book(data: "{\"name\": \"Painted House\",\"author\": \"bae-fd541c25-229e-5280-b44b-e5c2af3e374d\"}") {
						name
					}
				}`,
				ExpectedError: "no document for the given key exists",
			},
		},
	}
	simpleTests.ExecuteTestCase(t, test)
}

func TestMutationCreateOneToOne_UseAliasedRelationNameToLink_QueryFromPrimarySide(t *testing.T) {
	bookKey := "bae-3d236f89-6a31-5add-a36a-27971a2eac76"

	test := testUtils.TestCase{
		Description: "One to one create mutation with an alias relation.",
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
						create_Author(data: "{\"name\": \"John Grisham\",\"published\": \"%s\"}") {
							name
						}
					}`,
					bookKey,
				),
				Results: []map[string]any{
					{
						"name": "John Grisham",
					},
				},
			},
			testUtils.Request{
				Request: `query {
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
				Request: `query {
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

func TestMutationCreateOneToOne_UseAliasedRelationNameToLink_QueryFromSecondarySide(t *testing.T) {
	authorKey := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"

	test := testUtils.TestCase{
		Description: "One to one create mutation from secondary side with alias relation.",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
					create_Author(data: "{\"name\": \"John Grisham\"}") {
						_key
					}
				}`,
				Results: []map[string]any{
					{
						"_key": authorKey,
					},
				},
			},
			testUtils.Request{
				Request: fmt.Sprintf(
					`mutation {
						create_Book(data: "{\"name\": \"Painted House\",\"author\": \"%s\"}") {
							name
						}
					}`,
					authorKey,
				),
				Results: []map[string]any{
					{
						"name": "Painted House",
					},
				},
			},
			testUtils.Request{
				Request: `query {
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
			testUtils.Request{
				Request: `query {
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
