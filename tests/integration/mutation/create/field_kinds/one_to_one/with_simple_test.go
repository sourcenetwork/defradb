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
					"_publishedID": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
				}`,
				ExpectedError: "the given field does not exist. Name: notName",
			},
		},
	}
	executeTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestMutationCreateOneToOneNoChild(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"_publishedID": "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
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

func TestMutationCreateOneToOne(t *testing.T) {
	test := testUtils.TestCase{
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
					"_publishedID": testUtils.NewDocIndex(0, 0),
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

func TestMutationCreateOneToOneSecondarySide_CollectionApi(t *testing.T) {
	test := testUtils.TestCase{
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

func TestMutationCreateOneToOneSecondarySide_GQL(t *testing.T) {
	test := testUtils.TestCase{
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

func TestMutationCreateOneToOne_ErrorsGivenRelationAlreadyEstablishedViaPrimary(t *testing.T) {
	test := testUtils.TestCase{
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
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Saadi Shirazi",
					"_publishedID": testUtils.NewDocIndex(0, 0),
				},
				ExpectedError: "can not index a doc's field(s) that violates unique index",
			},
		},
	}

	executeTestCase(t, test)
}
