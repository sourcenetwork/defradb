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

func TestMutationCreateOneToOne_UseAliasWithInvalidField_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, alias relation, with an invalid field.",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"notName": "John Grisham",
					"published": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
				ExpectedError: "The given field does not exist. Name: notName",
			},
		},
	}
	executeTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationCreateOneToOne_UseAliasWithNonExistingRelationPrimarySide_CreatedDoc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, alias relation, from the wrong side",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"published": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author {
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

func TestMutationCreateOneToOne_UseAliasWithNonExistingRelationSecondarySide_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, alias relation, from the secondary side",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"author": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
				ExpectedError: "no document for the given ID exists",
			},
		},
	}
	executeTestCase(t, test)
}

func TestMutationCreateOneToOne_UseAliasedRelationNameToLink_QueryFromPrimarySide(t *testing.T) {
	bookID := "bae-3d236f89-6a31-5add-a36a-27971a2eac76"

	test := testUtils.TestCase{
		Description: "One to one create mutation with an alias relation.",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: fmt.Sprintf(
					`{
						"name": "John Grisham",
						"published": "%s"
					}`,
					bookID,
				),
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

	executeTestCase(t, test)
}

func TestMutationCreateOneToOne_UseAliasedRelationNameToLink_QueryFromSecondarySide(t *testing.T) {
	authorID := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"

	test := testUtils.TestCase{
		Description: "One to one create mutation from secondary side with alias relation.",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: fmt.Sprintf(
					`{
						"name": "Painted House",
						"author": "%s"
					}`,
					authorID,
				),
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

	executeTestCase(t, test)
}
