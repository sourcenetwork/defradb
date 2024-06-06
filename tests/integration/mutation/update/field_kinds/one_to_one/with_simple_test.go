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

	"github.com/sourcenetwork/immutable"
)

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationUpdateOneToOneNoChild(t *testing.T) {
	unknownID := "bae-be6d8024-4953-5a92-84b4-f042d25230c6"

	test := testUtils.TestCase{
		Description: "One to one create mutation, from the wrong side",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 1,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"published_id": "%s"
					}`,
					unknownID,
				),
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

func TestMutationUpdateOneToOne(t *testing.T) {
	bookID := "bae-dafb74e9-2bf1-5f12-aea9-967814592bad"

	test := testUtils.TestCase{
		Description: "One to one update mutation",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 1,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"published_id": "%s"
					}`,
					bookID,
				),
			},
			testUtils.Request{
				Request: `
					query {
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
				Request: `
					query {
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

func TestMutationUpdateOneToOneSecondarySide(t *testing.T) {
	authorID := "bae-42d197b8-d14f-5570-a55d-9e8714b2a82a"

	test := testUtils.TestCase{
		Description: "One to one create mutation, from the secondary side",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham"
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"author_id": "%s"
					}`,
					authorID,
				),
			},
			testUtils.Request{
				Request: `
					query {
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
				Request: `
					query {
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

func TestMutationUpdateOneToOne_RelationIDToLinkFromPrimarySide(t *testing.T) {
	author1ID := "bae-42d197b8-d14f-5570-a55d-9e8714b2a82a"
	bookID := "bae-dfce6a1a-27fa-5dde-bea7-44df2dffac1a"

	test := testUtils.TestCase{
		Description: "One to one update mutation using relation id from single side (wrong)",
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
						"author_id": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 1,
				DocID:        1,
				Doc: fmt.Sprintf(
					`{
						"published_id": "%s"
					}`,
					bookID,
				),
				ExpectedError: "target document is already linked to another document.",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToOne_RelationIDToLinkFromSecondarySide(t *testing.T) {
	author1ID := "bae-42d197b8-d14f-5570-a55d-9e8714b2a82a"
	author2ID := "bae-a34d8759-e549-5083-8ba6-e04038c41caa"

	test := testUtils.TestCase{
		Description: "One to one update mutation using relation id from secondary side",
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
						"author_id": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"author_id": "%s"
					}`,
					author2ID,
				),
				ExpectedError: "target document is already linked to another document.",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToOne_InvalidLengthRelationIDToLink_Error(t *testing.T) {
	author1ID := "bae-42d197b8-d14f-5570-a55d-9e8714b2a82a"
	invalidLenSubID := "35953ca-518d-9e6b-9ce6cd00eff5"
	invalidAuthorID := "bae-" + invalidLenSubID

	test := testUtils.TestCase{
		Description: "One to one update mutation using invalid relation id",
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
						"author_id": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"author_id": "%s"
					}`,
					invalidAuthorID,
				),
				ExpectedError: "uuid: incorrect UUID length 30 in string \"" + invalidLenSubID + "\"",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToOne_InvalidRelationIDToLinkFromSecondarySide_Error(t *testing.T) {
	author1ID := "bae-42d197b8-d14f-5570-a55d-9e8714b2a82a"
	invalidAuthorID := "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ee"

	test := testUtils.TestCase{
		Description: "One to one update mutation using relation id from secondary side",
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
						"author_id": "%s"
					}`,
					author1ID,
				),
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: fmt.Sprintf(
					`{
						"author_id": "%s"
					}`,
					invalidAuthorID,
				),
				ExpectedError: "document not found or not authorized to access",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationUpdateOneToOne_RelationIDToLinkFromSecondarySideWithWrongField_Error(t *testing.T) {
	author1ID := "bae-42d197b8-d14f-5570-a55d-9e8714b2a82a"
	author2ID := "bae-a34d8759-e549-5083-8ba6-e04038c41caa"

	test := testUtils.TestCase{
		Description: "One to one update mutation using relation id from secondary side, with a wrong field.",
		// This restiction is temporary due to a bug in the collection api, see
		// https://github.com/sourcenetwork/defradb/issues/1852 for more info.
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
						"author_id": "%s"
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
						"author_id": "%s"
					}`,
					author2ID,
				),
				ExpectedError: "In field \"notName\": Unknown field.",
			},
		},
	}

	executeTestCase(t, test)
}
