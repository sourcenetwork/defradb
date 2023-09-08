// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one

import (
	"fmt"
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
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
	executeTestCase(t, test)
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

	executeTestCase(t, test)
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
	executeTestCase(t, test)
}

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
				ExpectedError: "target document is already linked to another document.",
			},
		},
	}

	executeTestCase(t, test)
}

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
				ExpectedError: "target document is already linked to another document.",
			},
		},
	}

	executeTestCase(t, test)
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

	executeTestCase(t, test)
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

	executeTestCase(t, test)
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

	executeTestCase(t, test)
}
