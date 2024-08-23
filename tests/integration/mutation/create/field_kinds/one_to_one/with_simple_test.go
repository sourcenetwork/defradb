// Copyright 2022 Democratized Data Foundation
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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/sourcenetwork/immutable"
)

func TestMutationCreateOneToOne_WithInvalidField_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, with an invalid field.",
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
					"published_id": "bae-be6d8024-4953-5a92-84b4-f042d25230c6"
				}`,
				ExpectedError: "The given field does not exist. Name: notName",
			},
		},
	}
	executeTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationCreateOneToOneNoChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, from the wrong side",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"published_id": "bae-be6d8024-4953-5a92-84b4-f042d25230c6"
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

func TestMutationCreateOneToOne_NonExistingRelationSecondarySide_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, from the secondary side",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"author_id": "bae-be6d8024-4953-5a92-84b4-f042d25230c6"
				}`,
				ExpectedError: "document not found or not authorized to access",
			},
		},
	}
	executeTestCase(t, test)
}

func TestMutationCreateOneToOne(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation",
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
					"name":         "John Grisham",
					"published_id": testUtils.NewDocIndex(0, 0),
				},
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
				Request: `
					query {
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

func TestMutationCreateOneToOneSecondarySide(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation from secondary side",
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
					"name":      "Painted House",
					"author_id": testUtils.NewDocIndex(1, 0),
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

func TestMutationCreateOneToOne_ErrorsGivenRelationAlreadyEstablishedViaPrimary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, errors due to link already existing, primary side",
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
					"name":         "John Grisham",
					"published_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Saadi Shirazi",
					"published_id": testUtils.NewDocIndex(0, 0),
				},
				ExpectedError: "target document is already linked to another document.",
			},
		},
	}

	executeTestCase(t, test)
}

func TestMutationCreateOneToOne_ErrorsGivenRelationAlreadyEstablishedViaSecondary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One to one create mutation, errors due to link already existing, secondary side",
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
					"name":      "Painted House",
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Golestan",
					"author_id": testUtils.NewDocIndex(1, 0),
				},
				ExpectedError: "target document is already linked to another document.",
			},
		},
	}

	executeTestCase(t, test)
}
