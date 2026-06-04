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

func TestQueryOneToManyWithParentJoinGroupNumberAndNumberFilterOnJoin(t *testing.T) {
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
						Author (groupBy: [age]) {
							age
							GROUP {
								name
								published (filter: {rating: {_gt: 4.6}}) {
									name
									rating
								}
							}
						}
					}`,

				Results: map[string]any{
					"Author": []map[string]any{
						{
							"age": int64(327),
							"GROUP": []map[string]any{
								{
									"name": "Voltaire",
									"published": []map[string]any{
										{
											"name":   "Candide",
											"rating": 4.95,
										},
										{
											"name":   "Zadig",
											"rating": 4.91,
										},
									},
								},
								{
									"name":      "Simon Pelloutier",
									"published": []map[string]any{},
								},
							},
						},
						{
							"age": int64(65),
							"GROUP": []map[string]any{
								{
									"name": "John Grisham",
									"published": []map[string]any{
										{
											"name":   "Painted House",
											"rating": 4.9,
										},
									},
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithParentJoinGroupNumberAndNumberFilterOnGroup(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
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
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Client",
					"rating":    4.5,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Candide",
					"rating":    4.95,
					"_authorID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Zadig",
					"rating":    4.91,
					"_authorID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating":    2,
					"_authorID": testUtils.NewDocIndex(1, 2),
				},
			},
			&action.Request{
				Request: `query {
						Author (groupBy: [age]) {
							age
							GROUP (filter: {published: {rating: {_gt: 4.6}}}) {
								name
								published {
									name
									rating
								}
							}
						}
					}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"age": int64(327),
							"GROUP": []map[string]any{
								{
									"name": "Voltaire",
									"published": []map[string]any{
										{
											"name":   "Candide",
											"rating": 4.95,
										},
										{
											"name":   "Zadig",
											"rating": 4.91,
										},
									},
								},
							},
						},
						{
							"age": int64(65),
							"GROUP": []map[string]any{
								{
									"name": "John Grisham",
									"published": []map[string]any{
										{
											"name":   "Painted House",
											"rating": 4.9,
										},
										{
											"name":   "A Time for Mercy",
											"rating": 4.5,
										},
										{
											"name":   "The Client",
											"rating": 4.5,
										},
									},
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithParentJoinGroupNumberAndNumberFilterOnGroupAndOnGroupJoin(t *testing.T) {
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
						Author (groupBy: [age], filter: {age: {_gt: 300}}) {
							age
							GROUP {
								name
								published (filter: {rating: {_gt: 4.91}}){
									name
									rating
								}
							}
						}
					}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"age": int64(327),
							"GROUP": []map[string]any{
								{
									"name":      "Simon Pelloutier",
									"published": []map[string]any{},
								},
								{
									"name": "Voltaire",
									"published": []map[string]any{
										{
											"name":   "Candide",
											"rating": 4.95,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
