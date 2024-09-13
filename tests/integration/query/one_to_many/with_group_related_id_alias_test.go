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
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToManyWithParentGroupByOnRelatedTypeFromManySideUsingAlias(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many query with groupBy on related field alias (from many side).",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Client",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Candide",
					"rating":    4.95,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Zadig",
					"rating":    4.91,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating":    2,
					"author_id": testUtils.NewDocIndex(1, 2),
				},
			},
			testUtils.Request{
				Request: `query {
					Book(groupBy: [author]) {
						author_id
						_group {
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
							"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b",
							"_group": []map[string]any{
								{
									"name":   "Painted House",
									"rating": 4.9,
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
								{
									"name":   "A Time for Mercy",
									"rating": 4.5,
									"author": map[string]any{
										"age":  int64(65),
										"name": "John Grisham",
									},
								},
							},
						},
						{
							"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c",
							"_group": []map[string]any{
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
							"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614",
							"_group": []map[string]any{
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithParentGroupByOnRelatedTypeFromManySideUsingAliasAndRelatedSelection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many query with groupBy on related field alias (from many side).",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Client",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Candide",
					"rating":    4.95,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Zadig",
					"rating":    4.91,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating":    2,
					"author_id": testUtils.NewDocIndex(1, 2),
				},
			},
			testUtils.Request{
				Request: `query {
					Book(groupBy: [author]) {
						author {
							_docID
							name
						}
						_group {
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
								"name":   "John Grisham",
								"_docID": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b",
							},
							"_group": []map[string]any{
								{
									"name":   "Painted House",
									"rating": 4.9,
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
								{
									"name":   "A Time for Mercy",
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
								"name":   "Voltaire",
								"_docID": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c",
							},
							"_group": []map[string]any{
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
								"name":   "Simon Pelloutier",
								"_docID": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614",
							},
							"_group": []map[string]any{
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithParentGroupByOnRelatedTypeWithIDSelectionFromManySideUsingAlias(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many query with groupBy on related field alias, with id selection & related selection (from many side).",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Client",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Candide",
					"rating":    4.95,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Zadig",
					"rating":    4.91,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating":    2,
					"author_id": testUtils.NewDocIndex(1, 2),
				},
			},
			testUtils.Request{
				Request: `query {
					Book(groupBy: [author]) {
						author_id
						_group {
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
							"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b",
							"_group": []map[string]any{
								{
									"name":   "Painted House",
									"rating": 4.9,
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
								{
									"name":   "A Time for Mercy",
									"rating": 4.5,
									"author": map[string]any{
										"age":  int64(65),
										"name": "John Grisham",
									},
								},
							},
						},
						{
							"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c",
							"_group": []map[string]any{
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
							"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614",
							"_group": []map[string]any{
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithParentGroupByOnRelatedTypeWithIDSelectionFromManySideUsingAliasAndRelatedSelection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many query with groupBy on related field alias, with id selection & related selection (from many side).",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "The Client",
					"rating":    4.5,
					"author_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Candide",
					"rating":    4.95,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Zadig",
					"rating":    4.91,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating":    2,
					"author_id": testUtils.NewDocIndex(1, 2),
				},
			},
			testUtils.Request{
				Request: `query {
					Book(groupBy: [author]) {
						author_id
						author {
							_docID
							name
						}
						_group {
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
							"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b",
							"author": map[string]any{
								"name":   "John Grisham",
								"_docID": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b",
							},
							"_group": []map[string]any{
								{
									"name":   "Painted House",
									"rating": 4.9,
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
								{
									"name":   "A Time for Mercy",
									"rating": 4.5,
									"author": map[string]any{
										"age":  int64(65),
										"name": "John Grisham",
									},
								},
							},
						},
						{
							"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c",
							"author": map[string]any{
								"name":   "Voltaire",
								"_docID": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c",
							},
							"_group": []map[string]any{
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
							"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614",
							"author": map[string]any{
								"name":   "Simon Pelloutier",
								"_docID": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614",
							},
							"_group": []map[string]any{
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithParentGroupByOnRelatedTypeFromSingleSideUsingAlias(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many query with groupBy on related id field alias (from single side).",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614"
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
			testUtils.Request{
				Request: `query {
					Author(groupBy: [published]) {
						_group {
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
		Description: "One-to-many query with groupBy on related id field alias, with id selection (from single side).",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614"
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
			testUtils.Request{
				Request: `query {
					Author(groupBy: [published]) {
						published_id
						_group {
							name
						}
					}
				}`,
				ExpectedError: "Cannot query field \"published_id\" on type \"Author\". ",
			},
		},
	}

	executeTestCase(t, test)
}
