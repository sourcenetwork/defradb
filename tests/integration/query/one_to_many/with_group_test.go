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

func TestQueryOneToManyWithInnerJoinGroupNumber(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Painted House",
						"rating": 4.9,
						"_authorID": "{{.DocID1_0}}"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"_authorID": "{{.DocID1_0}}"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name": "The Client",
						"rating": 4.5,
						"_authorID": "{{.DocID1_0}}"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Theif Lord",
						"rating": 4.8,
						"_authorID": "{{.DocID1_1}}"
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
								GROUP {
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
									"GROUP": []map[string]any{
										{
											"name": "Painted House",
										},
									},
								},
								{
									"rating": 4.5,
									"GROUP": []map[string]any{
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
									"GROUP": []map[string]any{
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
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Painted House",
						"rating": 4.9,
						"_authorID": "{{.DocID1_0}}"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"_authorID": "{{.DocID1_0}}"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name": "The Client",
						"rating": 4.5,
						"_authorID": "{{.DocID1_0}}"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Candide",
						"rating": 4.95,
						"_authorID": "{{.DocID1_1}}"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Zadig",
						"rating": 4.91,
						"_authorID": "{{.DocID1_1}}"
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"_authorID": "{{.DocID1_2}}"
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

func TestQueryOneToManyWithParentGroupByOnRelationAndDuplicateRelationSelection(t *testing.T) {
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
				CollectionID: 0,
				Doc: `{
						"name": "Painted House",
						"rating": 4.9,
						"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
					}`,
			},
			&action.Request{
				// The relation `author` is both the group-by field and selected
				// twice at the parent level. The duplicated relation builds a
				// shared multiScanNode, which group expansion must not crash on.
				Request: `query {
					Book(groupBy: [author]) {
						author {
							name
						}
						author {
							name
						}
						GROUP {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"author": map[string]any{
								"name": "John Grisham",
							},
							"GROUP": []map[string]any{
								{
									"name": "Painted House",
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

func TestQueryOneToManyWithDuplicateRelationSelectionEachWithInnerGroupByOnRelation(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.ExplainRequest{
				Request: `query @explain(type: debug) {
					Author {
						published(groupBy: [author]) {
							author {
								name
							}
							GROUP {
								name
							}
						}
						published(groupBy: [author]) {
							author {
								name
							}
							GROUP {
								name
							}
						}
					}
				}`,
				ExpectedFullGraph: map[string]any{
					"explain": map[string]any{
						"operationNode": []map[string]any{
							{
								"selectTopNode": map[string]any{
									"selectNode": map[string]any{
										"typeIndexJoin": map[string]any{
											"typeJoinMany": map[string]any{
												"root": map[string]any{
													"scanNode": map[string]any{},
												},
												"subType": map[string]any{
													"selectTopNode": map[string]any{
														"groupNode": map[string]any{
															"selectNode": map[string]any{
																"pipeNode": map[string]any{
																	"typeIndexJoin": map[string]any{
																		"typeJoinOne": map[string]any{
																			"root": map[string]any{
																				"scanNode": map[string]any{},
																			},
																			"subType": map[string]any{
																				"selectTopNode": map[string]any{
																					"selectNode": map[string]any{
																						"scanNode": map[string]any{},
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
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
							GROUP {
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
