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
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationCreateOneToMany_AliasedRelationNameWithInvalidField_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many create mutation, with an invalid field, with alias.",
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GQL mutation will return a different error
			// when field types do not match
			testUtils.CollectionNamedMutationType,
			testUtils.CollectionSaveMutationType,
		}),
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"notName": "Painted House",
					"author": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
				ExpectedError: "The given field does not exist. Name: notName",
			},
		},
	}
	executeTestCase(t, test)
}

func TestMutationCreateOneToMany_AliasedRelationNameNonExistingRelationSingleSide_NoIDFieldError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many create mutation, non-existing id, from the single side, no id relation field, with alias.",
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GQL mutation will return a different error
			// when field types do not match
			testUtils.CollectionNamedMutationType,
			testUtils.CollectionSaveMutationType,
		}),
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John Grisham",
					"published": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
				ExpectedError: "The given field does not exist. Name: published",
			},
		},
	}
	executeTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationCreateOneToMany_AliasedRelationNameNonExistingRelationManySide_CreatedDoc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many create mutation, non-existing id, from the many side, with alias",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"author": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book {
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
	executeTestCase(t, test)
}
func TestMutationCreateOneToMany_AliasedRelationNameInvalidIDManySide_CreatedDoc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many create mutation, invalid id, from the many side, with alias",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"author": "ValueDoesntMatter"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book {
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
	executeTestCase(t, test)
}

func TestMutationCreateOneToMany_AliasedRelationNameToLinkFromManySide(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to many create mutation using relation id from many side, with alias.",
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

	executeTestCase(t, test)
}

func TestMutationUpdateOneToMany_AliasRelationNameAndInternalIDBothProduceSameDocID(t *testing.T) {
	// These IDs MUST be shared by both tests below.
	bookID := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	nonAliasedTest := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from single side (wrong)",
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
			},
			testUtils.Request{
				Request: `query {
					Book {
						_docID
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": bookID, // Must be same as below.
					},
				},
			},
		},
	}
	executeTestCase(t, nonAliasedTest)

	// Check that `bookID` is same in both above and the alised version below.
	// Note: Everything should be same, only diff should be the use of alias.

	aliasedTest := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from single side (wrong)",
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
			},
			testUtils.Request{
				Request: `query {
					Book {
						_docID
					}
				}`,
				Results: []map[string]any{
					{
						"_docID": bookID, // Must be same as below.
					},
				},
			},
		},
	}
	executeTestCase(t, aliasedTest)
}
