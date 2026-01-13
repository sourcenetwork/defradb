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
			testUtils.Request{
				Request: `query {
					Book {
						name
						_authorID
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name":      "Candide",
							"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32",
						},
						{
							"name":      "Painted House",
							"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab",
						},
						{
							"name":      "Zadig",
							"_authorID": "bae-b9c6cd5a-a931-5984-994d-7c435baa9f32",
						},
						{
							"name":      "A Time for Mercy",
							"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab",
						},
						{
							"name":      "The Client",
							"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab",
						},
						{
							"name":      "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
							"_authorID": "bae-7687d0c1-91b0-519e-99e4-eb92887663dd",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithRelatedTypeIDFromSingleSide(t *testing.T) {
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
			testUtils.Request{
				Request: `query {
					Author {
						name
						_authorID
					}
				}`,
				ExpectedError: "Cannot query field \"_authorID\" on type \"Author\".",
			},
		},
	}

	executeTestCase(t, test)
}
