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

// TODO: Don't return grouped field if not selected. [https://github.com/sourcenetwork/defradb/issues/1582].
func TestQueryOneToManyWithParentGroupByOnRelatedTypeIDFromManySide(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many query with groupBy on related id (from many side).",
		Request: `query {
				Book(groupBy: [author_id]) {
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
		Docs: map[int][]string{
			//books
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-7accaba8-ea9d-54b1-92f4-4a7ac5de88b3"
				}`,
				`{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-7accaba8-ea9d-54b1-92f4-4a7ac5de88b3"
				}`,
				`{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-09d33399-197a-5b98-b135-4398f2b6de4c"
				}`,
			},
			//authors
			1: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-7accaba8-ea9d-54b1-92f4-4a7ac5de88b3
				`{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
				// bae-09d33399-197a-5b98-b135-4398f2b6de4c
				`{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
		},
		Results: []map[string]any{
			{
				"author_id": "bae-7accaba8-ea9d-54b1-92f4-4a7ac5de88b3",
				"_group": []map[string]any{
					{
						"name":   "Candide",
						"rating": 4.95,
						"author": map[string]any{
							"age":  uint64(327),
							"name": "Voltaire",
						},
					},
					{
						"name":   "Zadig",
						"rating": 4.91,
						"author": map[string]any{
							"age":  uint64(327),
							"name": "Voltaire",
						},
					},
				},
			},
			{
				"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
				"_group": []map[string]any{
					{
						"name":   "The Client",
						"rating": 4.5,
						"author": map[string]any{
							"age":  uint64(65),
							"name": "John Grisham",
						},
					},
					{
						"name":   "Painted House",
						"rating": 4.9,
						"author": map[string]any{
							"age":  uint64(65),
							"name": "John Grisham",
						},
					},
					{
						"name":   "A Time for Mercy",
						"rating": 4.5,
						"author": map[string]any{
							"age":  uint64(65),
							"name": "John Grisham",
						},
					},
				},
			},
			{
				"author_id": "bae-09d33399-197a-5b98-b135-4398f2b6de4c",
				"_group": []map[string]any{
					{
						"name":   "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
						"rating": 2.0,
						"author": map[string]any{
							"age":  uint64(327),
							"name": "Simon Pelloutier",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithParentGroupByOnRelatedTypeIDWithIDSelectionFromManySide(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many query with groupBy on related id, with id selection (from many side).",
		Request: `query {
				Book(groupBy: [author_id]) {
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
		Docs: map[int][]string{
			//books
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-7accaba8-ea9d-54b1-92f4-4a7ac5de88b3"
				}`,
				`{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-7accaba8-ea9d-54b1-92f4-4a7ac5de88b3"
				}`,
				`{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-09d33399-197a-5b98-b135-4398f2b6de4c"
				}`,
			},
			//authors
			1: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-7accaba8-ea9d-54b1-92f4-4a7ac5de88b3
				`{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
				// bae-09d33399-197a-5b98-b135-4398f2b6de4c
				`{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
		},
		Results: []map[string]any{
			{
				"author_id": "bae-7accaba8-ea9d-54b1-92f4-4a7ac5de88b3",
				"_group": []map[string]any{
					{
						"name":   "Candide",
						"rating": 4.95,
						"author": map[string]any{
							"age":  uint64(327),
							"name": "Voltaire",
						},
					},
					{
						"name":   "Zadig",
						"rating": 4.91,
						"author": map[string]any{
							"age":  uint64(327),
							"name": "Voltaire",
						},
					},
				},
			},
			{
				"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3",
				"_group": []map[string]any{
					{
						"name":   "The Client",
						"rating": 4.5,
						"author": map[string]any{
							"age":  uint64(65),
							"name": "John Grisham",
						},
					},
					{
						"name":   "Painted House",
						"rating": 4.9,
						"author": map[string]any{
							"age":  uint64(65),
							"name": "John Grisham",
						},
					},
					{
						"name":   "A Time for Mercy",
						"rating": 4.5,
						"author": map[string]any{
							"age":  uint64(65),
							"name": "John Grisham",
						},
					},
				},
			},
			{
				"author_id": "bae-09d33399-197a-5b98-b135-4398f2b6de4c",
				"_group": []map[string]any{
					{
						"name":   "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
						"rating": 2.0,
						"author": map[string]any{
							"age":  uint64(327),
							"name": "Simon Pelloutier",
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

// TODO: Don't return grouped field if not selected. [https://github.com/sourcenetwork/defradb/issues/1582].
func TestQueryOneToManyWithParentGroupByOnRelatedTypeFromSingleSide(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many query with groupBy on related id (from single side).",
		Request: `query {
			Author(groupBy: [published_id]) {
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
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-7accaba8-ea9d-54b1-92f4-4a7ac5de88b3"
				}`,
				`{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-7accaba8-ea9d-54b1-92f4-4a7ac5de88b3"
				}`,
				`{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-09d33399-197a-5b98-b135-4398f2b6de4c"
				}`,
			},
			//authors
			1: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-7accaba8-ea9d-54b1-92f4-4a7ac5de88b3
				`{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
				// bae-09d33399-197a-5b98-b135-4398f2b6de4c
				`{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
		},

		ExpectedError: "Argument \"groupBy\" has invalid value [published_id].\nIn element #1: Expected type \"AuthorFields\", found published_id.",
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithParentGroupByOnRelatedTypeWithIDSelectionFromSingleSide(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many query with groupBy on related id, with id selection (from single side).",
		Request: `query {
			Author(groupBy: [published_id]) {
				published_id
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
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "A Time for Mercy",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "The Client",
					"rating": 4.5,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
				`{
					"name": "Candide",
					"rating": 4.95,
					"author_id": "bae-7accaba8-ea9d-54b1-92f4-4a7ac5de88b3"
				}`,
				`{
					"name": "Zadig",
					"rating": 4.91,
					"author_id": "bae-7accaba8-ea9d-54b1-92f4-4a7ac5de88b3"
				}`,
				`{
					"name": "Histoiare des Celtes et particulierement des Gaulois et des Germains depuis les temps fabuleux jusqua la prise de Roze par les Gaulois",
					"rating": 2,
					"author_id": "bae-09d33399-197a-5b98-b135-4398f2b6de4c"
				}`,
			},
			//authors
			1: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
				// bae-7accaba8-ea9d-54b1-92f4-4a7ac5de88b3
				`{
					"name": "Voltaire",
					"age": 327,
					"verified": true
				}`,
				// bae-09d33399-197a-5b98-b135-4398f2b6de4c
				`{
					"name": "Simon Pelloutier",
					"age": 327,
					"verified": true
				}`,
			},
		},

		ExpectedError: "Argument \"groupBy\" has invalid value [published_id].\nIn element #1: Expected type \"AuthorFields\", found published_id.",
	}

	executeTestCase(t, test)
}
