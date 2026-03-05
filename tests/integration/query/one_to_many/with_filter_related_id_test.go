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
)

func TestQueryFromManySideWithEqFilterOnRelatedType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Client",
					"rating": 4.5,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Candide",
					"rating": 4.95,
					"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zadig",
					"rating": 4.91,
					"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"_authorID": "bae-7687d0c1-91b0-519e-99e4-eb92887663dd"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
			&action.Request{
				Request: `query {
					Book(filter: {author: {_docID: {_eq: "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{"name": "Painted House"},
						{"name": "A Time for Mercy"},
						{"name": "The Client"},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryFromManySideWithFilterOnRelatedObjectID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Client",
					"rating": 4.5,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Candide",
					"rating": 4.95,
					"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zadig",
					"rating": 4.91,
					"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"_authorID": "bae-7687d0c1-91b0-519e-99e4-eb92887663dd"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
			&action.Request{
				Request: `query {
					Book(filter: {_authorID: {_eq: "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{"name": "Painted House"},
						{"name": "A Time for Mercy"},
						{"name": "The Client"},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryFromManySideWithSameFiltersInDifferentWayOnRelatedType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Client",
					"rating": 4.5,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Candide",
					"rating": 4.95,
					"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zadig",
					"rating": 4.91,
					"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"_authorID": "bae-7687d0c1-91b0-519e-99e4-eb92887663dd"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
			&action.Request{
				Request: `query {
					Book(
						filter: {
							author: {_docID: {_eq: "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"}},
							_authorID: {_eq: "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"}
						}
					) {
						name
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{"name": "Painted House"},
						{"name": "A Time for Mercy"},
						{"name": "The Client"},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryFromSingleSideWithEqFilterOnRelatedType(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Client",
					"rating": 4.5,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Candide",
					"rating": 4.95,
					"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zadig",
					"rating": 4.91,
					"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"_authorID": "bae-7687d0c1-91b0-519e-99e4-eb92887663dd"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
			&action.Request{
				Request: `query {
					Author(filter: {published: {_docID: {_eq: "bae-82bbdc18-aa15-57b8-83af-795a752b3b8f"}}}) {
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

func TestQueryFromSingleSideWithFilterOnRelatedObjectID_Error(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Client",
					"rating": 4.5,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Candide",
					"rating": 4.95,
					"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zadig",
					"rating": 4.91,
					"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32"
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"_authorID": "bae-7687d0c1-91b0-519e-99e4-eb92887663dd"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
			&action.Request{
				Request: `query {
					Author(filter: {_publishedID: {_eq: "bae-82bbdc18-aa15-57b8-83af-795a752b3b8f"}}) {
						name
					}
				}`,
				ExpectedError: "Argument \"filter\" has invalid value {_publishedID: {_eq: \"bae-82bbdc18-aa15-57b8-83af-795a752b3b8f\"}}.\nIn field \"_publishedID\": Unknown field.",
			},
		},
	}

	executeTestCase(t, test)
}
