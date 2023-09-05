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

func TestMutationUpdateOneToOne_AliasRelationNameToLinkFromPrimarySide(t *testing.T) {
	author1Key := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	author2Key := "bae-35953caf-4898-518d-9e6b-9ce6cd86ebe5"
	bookKey := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to one update mutation using alias relation id from single side",
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
 						create_Book(data: "{\"name\": \"Painted House\",\"author\": \"%s\"}") {
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
 						update_Author(id: "%s", data: "{\"published\": \"%s\"}") {
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

func TestMutationUpdateOneToOne_AliasRelationNameToLinkFromSecondarySide(t *testing.T) {
	author1Key := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	author2Key := "bae-35953caf-4898-518d-9e6b-9ce6cd86ebe5"
	bookKey := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to one update mutation using alias relation id from secondary side",
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
 						create_Book(data: "{\"name\": \"Painted House\",\"author\": \"%s\"}") {
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
 						update_Book(id: "%s", data: "{\"author\": \"%s\"}") {
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

func TestMutationUpdateOneToOne_AliasWithInvalidLengthRelationIDToLink_Error(t *testing.T) {
	author1Key := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	invalidLenSubKey := "35953ca-518d-9e6b-9ce6cd00eff5"
	invalidAuthorKey := "bae-" + invalidLenSubKey
	bookKey := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to one update mutation using invalid alias relation id",
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
 						create_Book(data: "{\"name\": \"Painted House\",\"author\": \"%s\"}") {
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
						update_Book(id: "%s", data: "{\"author\": \"%s\"}") {
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

func TestMutationUpdateOneToOne_InvalidAliasRelationNameToLinkFromSecondarySide_Error(t *testing.T) {
	author1Key := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	invalidAuthorKey := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ee"
	bookKey := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to one update mutation using alias relation id from secondary side",
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
 						create_Book(data: "{\"name\": \"Painted House\",\"author\": \"%s\"}") {
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
						update_Book(id: "%s", data: "{\"author\": \"%s\"}") {
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

func TestMutationUpdateOneToOne_AliasRelationNameToLinkFromSecondarySideWithWrongField_Error(t *testing.T) {
	author1Key := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	author2Key := "bae-35953caf-4898-518d-9e6b-9ce6cd86ebe5"
	bookKey := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to one update mutation using relation alias name from secondary side, with a wrong field.",
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
 						create_Book(data: "{\"name\": \"Painted House\",\"author\": \"%s\"}") {
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
 						update_Book(id: "%s", data: "{\"notName\": \"Unpainted Condo\",\"author\": \"%s\"}") {
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
