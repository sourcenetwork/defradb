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
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationCreateOneToOne_UseAliasWithInvalidField_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, alias relation, with an invalid field.",
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GQL mutation will return a different error
			// when field types do not match
			testUtils.CollectionNamedMutationType,
			testUtils.CollectionSaveMutationType,
		}),
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"notName": "John Grisham",
					"published": "bae-be6d8024-4953-5a92-84b4-f042d25230c6"
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
					"published": "bae-be6d8024-4953-5a92-84b4-f042d25230c6"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Author {
						name
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

func TestMutationCreateOneToOne_UseAliasedRelationNameToLink_QueryFromPrimarySide(t *testing.T) {
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
				DocMap: map[string]any{
					"name":      "John Grisham",
					"published": testUtils.NewDocIndex(0, 0),
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
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "Painted House",
							"author": map[string]any{
								"name": "John Grisham",
							},
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
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
							"published": map[string]any{
								"name": "Painted House",
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationCreateOneToOne_UseAliasedRelationNameToLink_CollectionAPI_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation from secondary side with alias relation.",
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			testUtils.CollectionSaveMutationType,
			testUtils.CollectionNamedMutationType,
		}),
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":   "Painted House",
					"author": testUtils.NewDocIndex(1, 0),
				},
				ExpectedError: "cannot set relation from secondary side",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationCreateOneToOne_UseAliasedRelationNameToLink_GQL_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation from secondary side with alias relation.",
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			testUtils.GQLRequestMutationType,
		}),
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":   "Painted House",
					"author": testUtils.NewDocIndex(1, 0),
				},
				ExpectedError: "Argument \"input\" has invalid value",
			},
		},
	}

	executeTestCase(t, test)
}
