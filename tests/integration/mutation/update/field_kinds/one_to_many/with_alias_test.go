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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpdateOneToMany_AliasRelationNameToLinkFromSingleSide_CollectionApi(t *testing.T) {
	author1ID := "bae-a47f80ab-1c30-53b3-9dac-04a4a3fda77e"
	bookID := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from single side (wrong)",
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
				CollectionID: 1,
				Doc: `{
					"name": "New Shahzad"
				}`,
			},
			testUtils.CreateDoc{
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
	author1ID := "bae-a47f80ab-1c30-53b3-9dac-04a4a3fda77e"
	bookID := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from single side (wrong)",
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
				CollectionID: 1,
				Doc: `{
					"name": "New Shahzad"
				}`,
			},
			testUtils.CreateDoc{
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
	author1ID := "bae-a47f80ab-1c30-53b3-9dac-04a4a3fda77e"
	invalidAuthorID := "bae-35953ca-518d-9e6b-9ce6cd00eff5"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from many side",
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
	author1ID := "bae-a47f80ab-1c30-53b3-9dac-04a4a3fda77e"
	invalidAuthorID := "bae-35953ca-518d-9e6b-9ce6cd00eff5"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from many side",
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
	author1ID := "bae-a47f80ab-1c30-53b3-9dac-04a4a3fda77e"
	author2ID := "bae-789d10d4-e54f-531b-ae81-e15100f8e506"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from many side, with a wrong field.",
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
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "New Shahzad"
				}`,
			},
			testUtils.CreateDoc{
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
	author1ID := "bae-a47f80ab-1c30-53b3-9dac-04a4a3fda77e"
	author2ID := "bae-789d10d4-e54f-531b-ae81-e15100f8e506"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from many side",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "New Shahzad"
				}`,
			},
			testUtils.CreateDoc{
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
							"name": "New Shahzad",
							"published": []map[string]any{
								{
									"name": "Painted House",
								},
							},
						},
						{
							"name":      "John Grisham",
							"published": []map[string]any{},
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
