// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Copyright 2023 Democratized Data Foundation
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
	"fmt"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	fixture "github.com/sourcenetwork/defradb/tests/integration/mutation/one_to_one"
)

// Note: This test should probably not pass, as even after updating a link to a new document
// from one side the previous link still remains on the other side of the link.
func TestMutationUpdateOneToOne_RelationIDToLinkFromPrimarySide(t *testing.T) {
	author1Key := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	author2Key := "bae-35953caf-4898-518d-9e6b-9ce6cd86ebe5"
	bookKey := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to one update mutation using relation id from single side (wrong)",
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
 						update_Author(id: "%s", data: "{\"published_id\": \"%s\"}") {
 							name
 						}
 					}`,
					author2Key,
					bookKey,
				),
				Results: []map[string]any{
					{
						"name": "New Shahzad",
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
					{
						"name": "New Shahzad",
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

	fixture.ExecuteTestCase(t, test)
}

// Note: This test should probably not pass, as even after updating a link to a new document
// from one side the previous link still remains on the other side of the link.
func TestMutationUpdateOneToOne_RelationIDToLinkFromSecondarySide(t *testing.T) {
	author1Key := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	author2Key := "bae-35953caf-4898-518d-9e6b-9ce6cd86ebe5"
	bookKey := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to one update mutation using relation id from secondary side",
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
						"name": "John Grisham",
						"published": map[string]any{
							"name": "Painted House",
						},
					},
					{
						"name": "New Shahzad",
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

	fixture.ExecuteTestCase(t, test)
}

func TestMutationUpdateOneToOne_InvalidLengthRelationIDToLink_Error(t *testing.T) {
	author1Key := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	invalidLenSubKey := "35953ca-518d-9e6b-9ce6cd00eff5"
	invalidAuthorKey := "bae-" + invalidLenSubKey
	bookKey := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to one update mutation using invalid relation id",
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
				ExpectedError: "uuid: incorrect UUID length 30 in string \"" + invalidLenSubKey + "\"",
			},
		},
	}

	fixture.ExecuteTestCase(t, test)
}

func TestMutationUpdateOneToOne_InvalidRelationIDToLinkFromSecondarySide_Error(t *testing.T) {
	author1Key := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	invalidAuthorKey := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ee"
	bookKey := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to one update mutation using relation id from secondary side",
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
				ExpectedError: "no document for the given key exists",
			},
		},
	}

	fixture.ExecuteTestCase(t, test)
}

func TestMutationUpdateOneToOne_RelationIDToLinkFromSecondarySideWithWrongField_Error(t *testing.T) {
	author1Key := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	author2Key := "bae-35953caf-4898-518d-9e6b-9ce6cd86ebe5"
	bookKey := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to one update mutation using relation id from secondary side, with a wrong field.",
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

	fixture.ExecuteTestCase(t, test)
}
