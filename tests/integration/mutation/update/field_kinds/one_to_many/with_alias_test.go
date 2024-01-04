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

func TestMutationUpdateOneToMany_AliasRelationNameToLinkFromSingleSide_Collection(t *testing.T) {
	author1ID := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	bookID := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from single side (wrong)",
		// This restiction is temporary due to an inconsitent error message, see
		// TestMutationUpdateOneToMany_AliasRelationNameToLinkFromSingleSide_GQL
		// and https://github.com/sourcenetwork/defradb/issues/1854 for more info.
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
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
				CollectionID: 1,
				DocID:        1,
				// NOTE: There is no `published` on book.
				Doc: fmt.Sprintf(
					`{
						"published": "%s"
					}`,
					bookID,
				),
				ExpectedError: "The given field does not exist. Name: published",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToMany_AliasRelationNameToLinkFromSingleSide_GQL(t *testing.T) {
	author1ID := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	bookID := "bae-22e0a1c2-d12b-5bfd-b039-0cf72f963991"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from single side (wrong)",
		// This restiction is temporary due to an inconsitent error message, see
		// TestMutationUpdateOneToMany_AliasRelationNameToLinkFromSingleSide_Collection
		// and https://github.com/sourcenetwork/defradb/issues/1854 for more info.
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
				// NOTE: There is no `published` on book.
				Doc: fmt.Sprintf(
					`{
						"published": "%s"
					}`,
					bookID,
				),
				ExpectedError: "The given field or alias to field does not exist. Name: published",
			},
		},
	}

	executeTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationUpdateOneToMany_InvalidAliasRelationNameToLinkFromManySide_GQL(t *testing.T) {
	author1ID := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	invalidAuthorID := "bae-35953ca-518d-9e6b-9ce6cd00eff5"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from many side",
		// This restiction is temporary due to a bug in the collection api, see
		// TestMutationUpdateOneToMany_InvalidAliasRelationNameToLinkFromManySide_Collection
		// and https://github.com/sourcenetwork/defradb/issues/1703 for more info.
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
						"name":      "John Grisham",
						"published": []map[string]any{},
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
						"name":   "Painted House",
						"author": nil, // Linked to incorrect id
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
//
// This test also documents a bug in the collection api, see:
// TestMutationUpdateOneToMany_InvalidAliasRelationNameToLinkFromManySide_GQL
// and https://github.com/sourcenetwork/defradb/issues/1703 for more info.
func TestMutationUpdateOneToMany_InvalidAliasRelationNameToLinkFromManySide_Collection(t *testing.T) {
	author1ID := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	invalidAuthorID := "bae-35953ca-518d-9e6b-9ce6cd00eff5"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from many side",
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
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
				ExpectedError: "The given field does not exist. Name: author",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToMany_AliasRelationNameToLinkFromManySideWithWrongField_Error(t *testing.T) {
	author1ID := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	author2ID := "bae-35953caf-4898-518d-9e6b-9ce6cd86ebe5"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from many side, with a wrong field.",
		// This restiction is temporary due to a bug in the collection api, see
		// TestMutationUpdateOneToMany_InvalidAliasRelationNameToLinkFromManySide_Collection
		// and https://github.com/sourcenetwork/defradb/issues/1703 for more info.
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
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"notName": "Unpainted Condo",
						"author": "%s"
					}`,
					author2ID,
				),
				ExpectedError: "The given field does not exist. Name: notName",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToMany_AliasRelationNameToLinkFromManySide(t *testing.T) {
	author1ID := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
	author2ID := "bae-35953caf-4898-518d-9e6b-9ce6cd86ebe5"

	test := testUtils.TestCase{
		Description: "One to many update mutation using relation alias name from many side",
		// This restiction is temporary due to a bug in the collection api, see
		// TestMutationUpdateOneToMany_InvalidAliasRelationNameToLinkFromManySide_Collection
		// and https://github.com/sourcenetwork/defradb/issues/1703 for more info.
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
				Results: []map[string]any{
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
							"name": "New Shahzad",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
