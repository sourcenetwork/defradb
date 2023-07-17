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
	fixture "github.com/sourcenetwork/defradb/tests/integration/mutation/one_to_many"
)

func TestMutationCreateOneToMany_AliasedRelationNameWithInvalidField_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many create mutation, with an invalid field, with alias.",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
					create_Book(data: "{\"notName\": \"Painted House\",\"author\": \"bae-fd541c25-229e-5280-b44b-e5c2af3e374d\"}") {
						name
					}
				}`,
				ExpectedError: "The given field does not exist. Name: notName",
			},
		},
	}
	fixture.ExecuteTestCase(t, test)
}

func TestMutationCreateOneToMany_AliasedRelationNameNonExistingRelationSingleSide_NoIDFieldError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many create mutation, non-existing id, from the single side, no id relation field, with alias.",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
					create_Author(data: "{\"name\": \"John Grisham\",\"published\": \"bae--b44b-e5c2af3e374d\"}") {
						name
					}
				}`,
				ExpectedError: "The given field does not exist. Name: published",
			},
		},
	}
	fixture.ExecuteTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationCreateOneToMany_AliasedRelationNameNonExistingRelationManySide_CreatedDoc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many create mutation, non-existing id, from the many side, with alias",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
					create_Book(data: "{\"name\": \"Painted House\",\"author\": \"bae-fd541c25-229e-5280-b44b-e5c2af3e374d\"}") {
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

func TestMutationCreateOneToMany_AliasedRelationNamToLinkFromSingleSide_NoIDFieldError(t *testing.T) {
	bookKey := "bae-3d236f89-6a31-5add-a36a-27971a2eac76"

	test := testUtils.TestCase{
		Description: "One to many create mutation with relation id from single side, with alias.",
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
				ExpectedError: "The given field does not exist. Name: published",
			},
		},
	}

	fixture.ExecuteTestCase(t, test)
}

func TestMutationCreateOneToMany_AliasedRelationNameToLinkFromManySide(t *testing.T) {
	authorKey := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"

	test := testUtils.TestCase{
		Description: "One to many create mutation using relation id from many side, with alias.",
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

func TestMutationUpdateOneToMany_AliasRelationNameAndInternalIDBothProduceSameDocID(t *testing.T) {
	// These keys MUST be shared by both tests below.
	authorKey := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	bookKey := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	nonAliasedTest := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from single side (wrong)",
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
 							_key
 							name
 						}
 					}`,
					authorKey,
				),
				Results: []map[string]any{
					{
						"_key": bookKey, // Must be same as below.
						"name": "Painted House",
					},
				},
			},
		},
	}
	fixture.ExecuteTestCase(t, nonAliasedTest)

	// Check that `bookKey` is same in both above and the alised version below.
	// Note: Everything should be same, only diff should be the use of alias.

	aliasedTest := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from single side (wrong)",
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
 							_key
 							name
 						}
 					}`,
					authorKey,
				),
				Results: []map[string]any{
					{
						"_key": bookKey, // Must be same as below.
						"name": "Painted House",
					},
				},
			},
		},
	}
	fixture.ExecuteTestCase(t, aliasedTest)
}
