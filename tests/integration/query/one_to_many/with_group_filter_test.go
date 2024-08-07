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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToManyWithParentJoinGroupNumberAndNumberFilterOnJoin(t *testing.T) {
	tests := []testUtils.TestCase{
		{
			Description: "One-to-many relation query from many side with parent level group and filter on join",
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
						Author (groupBy: [age]) {
							age
							_group {
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
										"name":      "Simon Pelloutier",
										"published": []map[string]any{},
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

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryOneToManyWithParentJoinGroupNumberAndNumberFilterOnGroup(t *testing.T) {
	tests := []testUtils.TestCase{
		{
			Description: "One-to-many relation query from many side with parent level group and group filter",
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
						Author (groupBy: [age]) {
							age
							_group (filter: {published: {rating: {_gt: 4.6}}}) {
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
								},
							},
							{
								"age": int64(65),
								"_group": []map[string]any{
									{
										"name": "John Grisham",
										"published": []map[string]any{
											{
												"name":   "The Client",
												"rating": 4.5,
											},
											{
												"name":   "Painted House",
												"rating": 4.9,
											},
											{
												"name":   "A Time for Mercy",
												"rating": 4.5,
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

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQueryOneToManyWithParentJoinGroupNumberAndNumberFilterOnGroupAndOnGroupJoin(t *testing.T) {
	tests := []testUtils.TestCase{
		{
			Description: "One-to-many relation query from many side with parent level group and filters",
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
						Author (groupBy: [age], filter: {age: {_gt: 300}}) {
							age
							_group {
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
								"_group": []map[string]any{
									{
										"name": "Voltaire",
										"published": []map[string]any{
											{
												"name":   "Candide",
												"rating": 4.95,
											},
										},
									},
									{
										"name":      "Simon Pelloutier",
										"published": []map[string]any{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}
