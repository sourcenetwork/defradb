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
	fixture "github.com/sourcenetwork/defradb/tests/integration/mutation/one_to_many"
)

func TestMutationCreateOneToMany_WithInvalidField_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many create mutation, with an invalid field.",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
					create_Book(data: "{\"notName\": \"Painted House\",\"author_id\": \"bae-fd541c25-229e-5280-b44b-e5c2af3e374d\"}") {
						name
					}
				}`,
				ExpectedError: "The given field does not exist. Name: notName",
			},
		},
	}
	fixture.ExecuteTestCase(t, test)
}

func TestMutationCreateOneToMany_NonExistingRelationSingleSide_NoIDFieldError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many create mutation, non-existing id, from the single side, no id relation field.",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
					create_Author(data: "{\"name\": \"John Grisham\",\"published_id\": \"bae--b44b-e5c2af3e374d\"}") {
						name
					}
				}`,
				ExpectedError: "The given field does not exist. Name: published_id",
			},
		},
	}
	fixture.ExecuteTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationCreateOneToMany_NonExistingRelationManySide_CreatedDoc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many create mutation, non-existing id, from the many side",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
					create_Book(data: "{\"name\": \"Painted House\",\"author_id\": \"bae-fd541c25-229e-5280-b44b-e5c2af3e374d\"}") {
						name
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Painted House",
					},
				},
			},
		},
	}
	fixture.ExecuteTestCase(t, test)
}

func TestMutationCreateOneToMany_RelationIDToLinkFromSingleSide_NoIDFieldError(t *testing.T) {
	bookKey := "bae-3d236f89-6a31-5add-a36a-27971a2eac76"

	test := testUtils.TestCase{
		Description: "One to many create mutation with relation id from single side.",
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
					bookKey,
				),
				ExpectedError: "The given field does not exist. Name: published_id",
			},
		},
	}

	fixture.ExecuteTestCase(t, test)
}

func TestMutationCreateOneToMany_RelationIDToLinkFromManySide(t *testing.T) {
	authorKey := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"

	test := testUtils.TestCase{
		Description: "One to many create mutation using relation id from many side",
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
						create_Book(data: "{\"name\": \"Painted House\",\"author_id\": \"%s\"}") {
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
						"published": []map[string]any{
							{
								"name": "Painted House",
							},
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

	fixture.ExecuteTestCase(t, test)
}
