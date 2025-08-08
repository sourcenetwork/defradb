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

func TestQueryOneToManyWithRelatedTypeIDFromManySide(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many query with related id (from many side).",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-01d16255-d8b0-53cd-9222-5237733e31d7"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-01d16255-d8b0-53cd-9222-5237733e31d7"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-c3b6ccf1-8f33-5259-a6d0-ae20594f03bf"
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
					Book {
						name
						author_id
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":      "Candide",
							"author_id": "bae-01d16255-d8b0-53cd-9222-5237733e31d7",
						},
						{
							"name":      "Painted House",
							"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80",
						},
						{
							"name":      "Zadig",
							"author_id": "bae-01d16255-d8b0-53cd-9222-5237733e31d7",
						},
						{
							"name":      "A Time for Mercy",
							"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80",
						},
						{
							"name":      "The Client",
							"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80",
						},
						{
							"name":      "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
							"author_id": "bae-c3b6ccf1-8f33-5259-a6d0-ae20594f03bf",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithRelatedTypeIDFromSingleSide(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many query with related id (from single side).",
		Actions: []any{
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-01d16255-d8b0-53cd-9222-5237733e31d7"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-01d16255-d8b0-53cd-9222-5237733e31d7"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-c3b6ccf1-8f33-5259-a6d0-ae20594f03bf"
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
					Author {
						name
						author_id
					}
				}`,
				ExpectedError: "Cannot query field \"author_id\" on type \"Author\".",
			},
		},
	}

	executeTestCase(t, test)
}
