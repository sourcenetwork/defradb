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
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"

	"github.com/sourcenetwork/immutable"
)

func TestMutationUpdateOneToMany_RelationIDToLinkFromSingleSide_Error(t *testing.T) {
	author1ID := "bae-5059e989-3cae-5584-9357-f3eb81e86241"
	bookID := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

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
					"name": "John Grisham"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "New Shahzad"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: fmt.Sprintf(
					`{
						"name": "Painted House",
						"_authorID": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 1,
				DocID:        1,
				// NOTE: There is no `_publishedID` on book.
				Doc: fmt.Sprintf(
					`{
						"_publishedID": "%s"
					}`,
					bookID,
				),
				ExpectedError: "the given field does not exist. Name: _publishedID",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToMany_InvalidRelationIDToLinkFromManySide(t *testing.T) {
	author1ID := "bae-5059e989-3cae-5584-9357-f3eb81e86241"
	invalidAuthorID := "bae-35953ca-518d-9e6b-9ce6cd00eff5"

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
				Doc: fmt.Sprintf(
					`{
						"name": "Painted House",
						"_authorID": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"_authorID": "%s"
					}`,
					invalidAuthorID,
				),
				ExpectedError: "uuid: incorrect UUID length 30 in string",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToMany_RelationIDToLinkFromManySideWithWrongField_Error(t *testing.T) {
	author1ID := "bae-5059e989-3cae-5584-9357-f3eb81e86241"
	author2ID := "bae-31e97109-6225-5be2-8c86-b16baa2782a3"

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
					"name": "John Grisham"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "New Shahzad"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: fmt.Sprintf(
					`{
						"name": "Painted House",
						"_authorID": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"notName": "Unpainted Condo",
						"_authorID": "%s"
					}`,
					author2ID,
				),
				ExpectedError: "the given field does not exist. Name: notName",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToMany_RelationIDToLinkFromManySide(t *testing.T) {
	author1ID := "bae-5059e989-3cae-5584-9357-f3eb81e86241"
	author2ID := "bae-31e97109-6225-5be2-8c86-b16baa2782a3"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "New Shahzad"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: fmt.Sprintf(
					`{
						"name": "Painted House",
						"_authorID": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"_authorID": "%s"
					}`,
					author2ID,
				),
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
				NonOrderedResults: true,
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
								"name": "New Shahzad",
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
