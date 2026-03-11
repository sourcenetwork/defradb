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

package one_to_one

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestMutationAddOneToOne_UseAliasWithInvalidField_Error(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			// GQL mutation will return a different error
			// when field types do not match
			state.CollectionNamedMutationType,
			state.CollectionSaveMutationType,
		}),
		Actions: []any{
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"notName": "John Grisham",
					"published": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
				}`,
				ExpectedError: "the given field does not exist. Name: notName",
			},
		},
	}
	executeTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationAddOneToOne_UseAliasWithNonExistingRelationPrimarySide_AddedDoc(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"published": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
				}`,
			},
			&action.Request{
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

func TestMutationAddOneToOne_UseAliasedRelationNameToLink_QueryFromPrimarySide(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":      "John Grisham",
					"published": testUtils.NewDocIndex(0, 0),
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

func TestMutationAddOneToOne_UseAliasedRelationNameToLink_CollectionAPI_Errors(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			state.CollectionSaveMutationType,
			state.CollectionNamedMutationType,
		}),
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
				ExpectedError: "cannot set relation from secondary side",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationAddOneToOne_UseAliasedRelationNameToLink_GQL_Errors(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			state.GQLRequestMutationType,
		}),
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
				ExpectedError: "Argument \"input\" has invalid value",
			},
		},
	}

	executeTestCase(t, test)
}
