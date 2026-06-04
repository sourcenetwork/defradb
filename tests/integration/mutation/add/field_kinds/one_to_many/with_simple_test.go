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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"

	"github.com/sourcenetwork/immutable"
)

func TestMutationAddOneToMany_WithInvalidField_Error(t *testing.T) {
	test := testUtils.TestCase{
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
					"_authorID": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
				}`,
				ExpectedError: "the given field does not exist. Name: notName",
			},
		},
	}
	executeTestCase(t, test)
}

func TestMutationAddOneToMany_NonExistingRelationSingleSide_NoIDFieldError(t *testing.T) {
	test := testUtils.TestCase{
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
					"_publishedID": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
				}`,
				ExpectedError: "the given field does not exist. Name: _publishedID",
			},
		},
	}
	executeTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationAddOneToMany_NonExistingRelationManySide_AddedDoc(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"_authorID": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
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

func TestMutationAddOneToMany_RelationIDToLinkFromManySide(t *testing.T) {
	test := testUtils.TestCase{
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
					"name":      "Painted House",
					"_authorID": testUtils.NewDocIndex(1, 0),
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
