// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package one_to_many

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestMutationAddOneToMany_AliasedRelationNameWithInvalidField_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a document with an invalid field name returns an error.",
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			// GQL mutation will return a different error
			// when field types do not match
			state.CollectionNamedMutationType,
			state.CollectionSaveMutationType,
		}),
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"notName": "Painted House",
					"author": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
				}`,
				ExpectedError: "the given field does not exist. Name: notName",
			},
		},
	}
	executeTestCase(t, test)
}

func TestMutationAddOneToMany_AliasedRelationNameNonExistingRelationSingleSide_NoIDFieldError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a document using the aliased relation field on the one-side returns an error.",
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			// GQL mutation will return a different error
			// when field types do not match
			state.CollectionNamedMutationType,
			state.CollectionSaveMutationType,
		}),
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John Grisham",
					"published": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
				}`,
				ExpectedError: "the given field does not exist. Name: published",
			},
		},
	}
	executeTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationAddOneToMany_AliasedRelationNameNonExistingRelationManySide_AddedDoc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a document with an aliased relation referencing a non-existing document succeeds.",
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"author": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
				}`,
			},
			&action.Request{
				Request: `query {
					Book {
						name
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "Painted House",
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

func TestMutationAddOneToMany_AliasedRelationNameToLinkFromManySide(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding a document using the aliased relation name correctly links the one-to-many relation.",
		Actions: []any{
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":   "Painted House",
					"author": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.Request{
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
							"published": []map[string]any{
								{
									"name": "Painted House",
								},
							},
						},
					},
				},
			},
			&action.Request{
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
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToMany_AliasRelationNameAndInternalIDBothProduceSameDocID(t *testing.T) {
	// These IDs MUST be shared by both tests below.
	bookID := "bae-a2df247a-8bc2-5761-9557-90400f490eef"

	nonAliasedTest := testUtils.TestCase{
		Description: "Linking via internal relation ID field produces the same docID as using the aliased name.",
		Actions: []any{
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":   "Painted House",
					"author": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.Request{
				Request: `query {
					Book {
						_docID
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"_docID": bookID, // Must be same as below.
						},
					},
				},
			},
		},
	}
	executeTestCase(t, nonAliasedTest)

	// Check that `bookID` is same in both above and the alised version below.
	// Note: Everything should be same, only diff should be the use of alias.

	aliasedTest := testUtils.TestCase{
		Description: "Linking via aliased relation name produces the same docID as the internal relation ID field.",
		Actions: []any{
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":   "Painted House",
					"author": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.Request{
				Request: `query {
					Book {
						_docID
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"_docID": bookID, // Must be same as below.
						},
					},
				},
			},
		},
	}
	executeTestCase(t, aliasedTest)
}
