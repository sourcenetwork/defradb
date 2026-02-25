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
	"fmt"
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestMutationUpdateOneToMany_AliasRelationNameToLinkFromSingleSide_CollectionApi(t *testing.T) {
	author1ID := "bae-5059e989-3cae-5584-9357-f3eb81e86241"
	bookID := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

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
						"author": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 1,
				DocID:        1,
				Doc: fmt.Sprintf(
					`{
						"published": "%s"
					}`,
					bookID,
				),
				ExpectedError: "cannot set relation from secondary side",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToMany_AliasRelationNameToLinkFromSingleSide_GQL(t *testing.T) {
	author1ID := "bae-5059e989-3cae-5584-9357-f3eb81e86241"
	bookID := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

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
						"author": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 1,
				DocID:        1,
				Doc: fmt.Sprintf(
					`{
						"published": "%s"
					}`,
					bookID,
				),
				ExpectedError: "Argument \"input\" has invalid value",
			},
		},
	}

	executeTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationUpdateOneToMany_InvalidAliasRelationNameToLinkFromManySide_GQL(t *testing.T) {
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
						"author": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"author": "%s"
					}`,
					invalidAuthorID,
				),
				ExpectedError: "uuid: incorrect UUID length 30 in string",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToMany_InvalidAliasRelationNameToLinkFromManySide_Collection(t *testing.T) {
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
						"author": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"author": "%s"
					}`,
					invalidAuthorID,
				),
				ExpectedError: "uuid: incorrect UUID length 30 in string",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToMany_AliasRelationNameToLinkFromManySideWithWrongField_Error(t *testing.T) {
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
						"author": "%s"
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
						"author": "%s"
					}`,
					author2ID,
				),
				ExpectedError: "the given field does not exist. Name: notName",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToMany_AliasRelationNameToLinkFromManySide(t *testing.T) {
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
						"author": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"author": "%s"
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
