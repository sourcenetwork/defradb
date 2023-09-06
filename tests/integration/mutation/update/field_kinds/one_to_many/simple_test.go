// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many

import (
	"fmt"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdateOneToMany_RelationIDToLinkFromSingleSide_Error(t *testing.T) {
	author1Key := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	author2Key := "bae-35953caf-4898-518d-9e6b-9ce6cd86ebe5"
	bookKey := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation id from single side (wrong)",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
 					create_Author(data: "{\"name\": \"John Grisham\"}") {
 						_key
 					}
 				}`,
				Results: []map[string]any{
					{
						"_key": author1Key,
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
 					create_Author(data: "{\"name\": \"New Shahzad\"}") {
 						_key
 					}
 				}`,
				Results: []map[string]any{
					{
						"_key": author2Key,
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
					author1Key,
				),
				Results: []map[string]any{
					{
						"_key": bookKey,
						"name": "Painted House",
					},
				},
			},
			testUtils.Request{ // NOTE: There is no `published_id` on book.
				Request: fmt.Sprintf(
					`mutation {
 						update_Author(id: "%s", data: "{\"published_id\": \"%s\"}") {
 							name
 						}
 					}`,
					author2Key,
					bookKey,
				),
				ExpectedError: "The given field does not exist. Name: published_id",
			},
		},
	}

	executeTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationUpdateOneToMany_InvalidRelationIDToLinkFromManySide(t *testing.T) {
	author1Key := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	invalidAuthorKey := "bae-35953ca-518d-9e6b-9ce6cd00eff5"
	bookKey := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation id from many side",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
 					create_Author(data: "{\"name\": \"John Grisham\"}") {
 						_key
 					}
 				}`,
				Results: []map[string]any{
					{
						"_key": author1Key,
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
					author1Key,
				),
				Results: []map[string]any{
					{
						"_key": bookKey,
						"name": "Painted House",
					},
				},
			},
			testUtils.Request{
				Request: fmt.Sprintf(
					`mutation {
 						update_Book(id: "%s", data: "{\"author_id\": \"%s\"}") {
 							name
 						}
 					}`,
					bookKey,
					invalidAuthorKey,
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
						"name":      "John Grisham",
						"published": []map[string]any{},
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
						"name":   "Painted House",
						"author": nil, // Linked to incorrect id
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToMany_RelationIDToLinkFromManySideWithWrongField_Error(t *testing.T) {
	author1Key := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	author2Key := "bae-35953caf-4898-518d-9e6b-9ce6cd86ebe5"
	bookKey := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation id from many side, with a wrong field.",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
 					create_Author(data: "{\"name\": \"John Grisham\"}") {
 						_key
 					}
 				}`,
				Results: []map[string]any{
					{
						"_key": author1Key,
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
 					create_Author(data: "{\"name\": \"New Shahzad\"}") {
 						_key
 					}
 				}`,
				Results: []map[string]any{
					{
						"_key": author2Key,
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
					author1Key,
				),
				Results: []map[string]any{
					{
						"_key": bookKey,
						"name": "Painted House",
					},
				},
			},
			testUtils.Request{
				Request: fmt.Sprintf(
					`mutation {
 						update_Book(id: "%s", data: "{\"notName\": \"Unpainted Condo\",\"author_id\": \"%s\"}") {
 							name
 						}
 					}`,
					bookKey,
					author2Key,
				),
				ExpectedError: "The given field does not exist. Name: notName",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToMany_RelationIDToLinkFromManySide(t *testing.T) {
	author1Key := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	author2Key := "bae-35953caf-4898-518d-9e6b-9ce6cd86ebe5"
	bookKey := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation id from many side",
		Actions: []any{
			testUtils.Request{
				Request: `mutation {
 					create_Author(data: "{\"name\": \"John Grisham\"}") {
 						_key
 					}
 				}`,
				Results: []map[string]any{
					{
						"_key": author1Key,
					},
				},
			},
			testUtils.Request{
				Request: `mutation {
 					create_Author(data: "{\"name\": \"New Shahzad\"}") {
 						_key
 					}
 				}`,
				Results: []map[string]any{
					{
						"_key": author2Key,
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
					author1Key,
				),
				Results: []map[string]any{
					{
						"_key": bookKey,
						"name": "Painted House",
					},
				},
			},
			testUtils.Request{
				Request: fmt.Sprintf(
					`mutation {
 						update_Book(id: "%s", data: "{\"author_id\": \"%s\"}") {
 							name
 						}
 					}`,
					bookKey,
					author2Key,
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
						"name":      "John Grisham",
						"published": []map[string]any{},
					},
					{
						"name": "New Shahzad",
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
							"name": "New Shahzad",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
