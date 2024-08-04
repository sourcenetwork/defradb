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

func TestQueryOneToManyWithInnerJoinGroupNumber(t *testing.T) {
	tests := []testUtils.RequestTestCase{
		{
			Description: "One-to-many relation query from many side with group inside of join",
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
			Docs: map[int][]string{
				//books
				0: { // bae-be6d8024-4953-5a92-84b4-f042d25230c6
					`{
						"name": "Painted House",
						"rating": 4.9,
						"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
					}`,
					`{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
					}`,
					`{
						"name": "The Client",
						"rating": 4.5,
						"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
					}`,
					`{
						"name": "Theif Lord",
						"rating": 4.8,
						"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
					}`,
				},
				//authors
				1: {
					// bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b
					`{
						"name": "John Grisham",
						"age": 65,
						"verified": true
					}`,
					// bae-72e8c691-9f20-55e7-9228-8af1cf54cace
					`{
						"name": "Cornelia Funke",
						"age": 62,
						"verified": false
					}`,
				},
			},
			Results: map[string]any{
				"Author": []map[string]any{
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
					{
						"name": "John Grisham",
						"age":  int64(65),
						"published": []map[string]any{
							{
								"rating": 4.5,
								"_group": []map[string]any{
									{
										"name": "The Client",
									},
									{
										"name": "A Time for Mercy",
									},
								},
							},
							{
								"rating": 4.9,
								"_group": []map[string]any{
									{
										"name": "Painted House",
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

func TestQueryOneToManyWithParentJoinGroupNumber(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from many side with parent level group",
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
		Docs: map[int][]string{
			//books
			0: { // bae-be6d8024-4953-5a92-84b4-f042d25230c6
				`{
						"name": "Painted House",
						"rating": 4.9,
						"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
					}`,
				`{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
					}`,
				`{
						"name": "The Client",
						"rating": 4.5,
						"author_id": "bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b"
					}`,
				`{
						"name": "Candide",
						"rating": 4.95,
						"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
					}`,
				`{
						"name": "Zadig",
						"rating": 4.91,
						"author_id": "bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c"
					}`,
				`{
						"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
						"rating": 2,
						"author_id": "bae-34a9bd41-1f0d-5748-8446-48fc36ef2614"
					}`,
			},
			//authors
			1: {
				// bae-e1ea288f-09fa-55fa-b0b5-0ac8941ea35b
				`{
						"name": "John Grisham",
						"age": 65,
						"verified": true
					}`,
				// bae-1594d2aa-d63c-51d2-8e5e-06ee0c9e2e8c
				`{
						"name": "Voltaire",
						"age": 327,
						"verified": true
					}`,
				// bae-34a9bd41-1f0d-5748-8446-48fc36ef2614
				`{
						"name": "Simon Pelloutier",
						"age": 327,
						"verified": true
					}`,
			},
		},
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
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithInnerJoinGroupNumberWithNonGroupFieldsSelected(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from many side with group inside of join and invalid field",
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
		Docs:          map[int][]string{},
		ExpectedError: "cannot select a non-group-by field at group-level",
	}

	executeTestCase(t, test)
}
