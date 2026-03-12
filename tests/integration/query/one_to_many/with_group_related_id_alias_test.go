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

func TestQueryOneToManyWithParentGroupByOnRelatedTypeFromManySideUsingAlias(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						published: [Book]
					}
				`,
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
					Book(groupBy: [author]) {
						_authorID
						GROUP {
							name
							rating
							author {
								name
								age
							}
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32",
							"GROUP": []map[string]any{
								{
									"name":   "Candide",
									"rating": 4.95,
									"author": map[string]any{
										"age":  int64(327),
										"name": "Voltaire",
									},
								},
								{
									"name":   "Zadig",
									"rating": 4.91,
									"author": map[string]any{
										"age":  int64(327),
										"name": "Voltaire",
									},
								},
							},
						},
						{
							"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab",
							"GROUP": []map[string]any{
								{
									"name":   "Painted House",
									"rating": 4.9,
									"author": map[string]any{
										"age":  int64(65),
										"name": "John Grisham",
									},
								},
								{
									"name":   "A Time for Mercy",
									"rating": 4.5,
									"author": map[string]any{
										"age":  int64(65),
										"name": "John Grisham",
									},
								},
								{
									"name":   "The Client",
									"rating": 4.5,
									"author": map[string]any{
										"age":  int64(65),
										"name": "John Grisham",
									},
								},
							},
						},
						{
							"_authorID": "bae-7687d0c1-91b0-519e-99e4-eb92887663dd",
							"GROUP": []map[string]any{
								{
									"name":   "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
									"rating": 2.0,
									"author": map[string]any{
										"age":  int64(327),
										"name": "Simon Pelloutier",
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

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithParentGroupByOnRelatedTypeFromManySideUsingAliasAndRelatedSelection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						published: [Book]
					}
				`,
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
					Book(groupBy: [author]) {
						author {
							_docID
							name
						}
						GROUP {
							name
							rating
							author {
								name
								age
							}
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"author": map[string]any{
								"name":   "Voltaire",
								"_docID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32",
							},
							"GROUP": []map[string]any{
								{
									"name":   "Candide",
									"rating": 4.95,
									"author": map[string]any{
										"age":  int64(327),
										"name": "Voltaire",
									},
								},
								{
									"name":   "Zadig",
									"rating": 4.91,
									"author": map[string]any{
										"age":  int64(327),
										"name": "Voltaire",
									},
								},
							},
						},
						{
							"author": map[string]any{
								"name":   "John Grisham",
								"_docID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab",
							},
							"GROUP": []map[string]any{
								{
									"name":   "Painted House",
									"rating": 4.9,
									"author": map[string]any{
										"age":  int64(65),
										"name": "John Grisham",
									},
								},
								{
									"name":   "A Time for Mercy",
									"rating": 4.5,
									"author": map[string]any{
										"age":  int64(65),
										"name": "John Grisham",
									},
								},
								{
									"name":   "The Client",
									"rating": 4.5,
									"author": map[string]any{
										"age":  int64(65),
										"name": "John Grisham",
									},
								},
							},
						},
						{
							"author": map[string]any{
								"name":   "Simon Pelloutier",
								"_docID": "bae-7687d0c1-91b0-519e-99e4-eb92887663dd",
							},
							"GROUP": []map[string]any{
								{
									"name":   "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
									"rating": 2.0,
									"author": map[string]any{
										"age":  int64(327),
										"name": "Simon Pelloutier",
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

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithParentGroupByOnRelatedTypeWithIDSelectionFromManySideUsingAlias(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						published: [Book]
					}
				`,
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
					Book(groupBy: [author]) {
						_authorID
						GROUP {
							name
							rating
							author {
								name
								age
							}
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32",
							"GROUP": []map[string]any{
								{
									"name":   "Candide",
									"rating": 4.95,
									"author": map[string]any{
										"age":  int64(327),
										"name": "Voltaire",
									},
								},
								{
									"name":   "Zadig",
									"rating": 4.91,
									"author": map[string]any{
										"age":  int64(327),
										"name": "Voltaire",
									},
								},
							},
						},
						{
							"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab",
							"GROUP": []map[string]any{
								{
									"name":   "Painted House",
									"rating": 4.9,
									"author": map[string]any{
										"age":  int64(65),
										"name": "John Grisham",
									},
								},
								{
									"name":   "A Time for Mercy",
									"rating": 4.5,
									"author": map[string]any{
										"age":  int64(65),
										"name": "John Grisham",
									},
								},
								{
									"name":   "The Client",
									"rating": 4.5,
									"author": map[string]any{
										"age":  int64(65),
										"name": "John Grisham",
									},
								},
							},
						},
						{
							"_authorID": "bae-7687d0c1-91b0-519e-99e4-eb92887663dd",
							"GROUP": []map[string]any{
								{
									"name":   "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
									"rating": 2.0,
									"author": map[string]any{
										"age":  int64(327),
										"name": "Simon Pelloutier",
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

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithParentGroupByOnRelatedTypeWithIDSelectionFromManySideUsingAliasAndRelatedSelection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						rating: Float
						author: Author
					}

					type Author {
						name: String
						age: Int
						verified: Boolean
						published: [Book]
					}
				`,
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
					Book(groupBy: [author]) {
						_authorID
						author {
							_docID
							name
						}
						GROUP {
							name
							rating
							author {
								name
								age
							}
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32",
							"author": map[string]any{
								"name":   "Voltaire",
								"_docID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32",
							},
							"GROUP": []map[string]any{
								{
									"name":   "Candide",
									"rating": 4.95,
									"author": map[string]any{
										"age":  int64(327),
										"name": "Voltaire",
									},
								},
								{
									"name":   "Zadig",
									"rating": 4.91,
									"author": map[string]any{
										"age":  int64(327),
										"name": "Voltaire",
									},
								},
							},
						},
						{
							"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab",
							"author": map[string]any{
								"name":   "John Grisham",
								"_docID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab",
							},
							"GROUP": []map[string]any{
								{
									"name":   "Painted House",
									"rating": 4.9,
									"author": map[string]any{
										"age":  int64(65),
										"name": "John Grisham",
									},
								},
								{
									"name":   "A Time for Mercy",
									"rating": 4.5,
									"author": map[string]any{
										"age":  int64(65),
										"name": "John Grisham",
									},
								},
								{
									"name":   "The Client",
									"rating": 4.5,
									"author": map[string]any{
										"age":  int64(65),
										"name": "John Grisham",
									},
								},
							},
						},
						{
							"_authorID": "bae-7687d0c1-91b0-519e-99e4-eb92887663dd",
							"author": map[string]any{
								"name":   "Simon Pelloutier",
								"_docID": "bae-7687d0c1-91b0-519e-99e4-eb92887663dd",
							},
							"GROUP": []map[string]any{
								{
									"name":   "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
									"rating": 2.0,
									"author": map[string]any{
										"age":  int64(327),
										"name": "Simon Pelloutier",
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

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithParentGroupByOnRelatedTypeFromSingleSideUsingAlias(t *testing.T) {
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
					Author(groupBy: [published]) {
						GROUP {
							name
						}
					}
				}`,
				ExpectedError: "invalid field value to groupBy. Field: published",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithParentGroupByOnRelatedTypeWithIDSelectionFromSingleSideUsingAlias(t *testing.T) {
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
					Author(groupBy: [published]) {
						_publishedID
						GROUP {
							name
						}
					}
				}`,
				ExpectedError: "Cannot query field \"_publishedID\" on type \"Author\". ",
			},
		},
	}

	executeTestCase(t, test)
}
