// Copyright 2022 Democratized Data Foundation
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
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToManyWithInnerJoinGroupNumber(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Painted House",
						"rating": 4.9,
						"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name": "The Client",
						"rating": 4.5,
						"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Theif Lord",
						"rating": 4.8,
						"_authorID": "bae-3d5a3204-4e55-5236-992a-ce27da27902b"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
						"name": "John Grisham",
						"age": 65,
						"verified": true
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
						"name": "Cornelia Funke",
						"age": 62,
						"verified": false
					}`,
			},
			&action.Request{
				Request: `query {
						Author {
							name
							age
							published (groupBy: [rating]){
								rating
								_group {
									name
								}
							}
						}
					}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "John Grisham",
							"age":  int64(65),
							"published": []map[string]any{
								{
									"rating": 4.9,
									"_group": []map[string]any{
										{
											"name": "Painted House",
										},
									},
								},
								{
									"rating": 4.5,
									"_group": []map[string]any{
										{
											"name": "A Time for Mercy",
										},
										{
											"name": "The Client",
										},
									},
								},
							},
						},
						{
							"name": "Cornelia Funke",
							"age":  int64(62),
							"published": []map[string]any{
								{
									"rating": 4.8,
									"_group": []map[string]any{
										{
											"name": "Theif Lord",
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

func TestQueryOneToManyWithParentJoinGroupNumber(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Painted House",
						"rating": 4.9,
						"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name": "The Client",
						"rating": 4.5,
						"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Candide",
						"rating": 4.95,
						"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Zadig",
						"rating": 4.91,
						"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"_authorID": "bae-7687d0c1-91b0-519e-99e4-eb92887663dd"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
						"name": "John Grisham",
						"age": 65,
						"verified": true
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
						"name": "Voltaire",
						"age": 327,
						"verified": true
					}`,
			},
			testUtils.CreateDoc{
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
						_group {
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
							"_group": []map[string]any{
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
									"name": "Simon Pelloutier",
									"published": []map[string]any{
										{
											"name":   "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
											"rating": float64(2),
										},
									},
								},
							},
						},
						{
							"age": int64(65),
							"_group": []map[string]any{
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

func TestQueryOneToManyWithInnerJoinGroupNumberWithNonGroupFieldsSelected(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					Author {
						name
						age
						published (groupBy: [rating]){
							rating
							name
							_group {
								name
							}
						}
					}
				}`,
				ExpectedError: "cannot select a non-group-by field at group-level",
			},
		},
	}

	executeTestCase(t, test)
}
