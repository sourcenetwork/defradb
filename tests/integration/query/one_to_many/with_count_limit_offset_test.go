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

func TestQueryOneToManyWithCountAndLimitAndOffset(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with count and limit and offset",
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
					"name": "The Firm",
					"rating": 4.1,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Pelican Brief",
					"rating": 4.0,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-f62bb529-3508-529d-8098-f97f9b67824c"
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
			testUtils.Request{
				Request: `query {
					Author {
						name
						_count(published: {})
						published(limit: 2, offset: 1) {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":   "John Grisham",
							"_count": 4,
							"published": []map[string]any{
								{
									"name": "Painted House",
								},
								{
									"name": "A Time for Mercy",
								},
							},
						},
						{
							"name":      "Cornelia Funke",
							"_count":    1,
							"published": []map[string]any{},
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithCountAndDifferentOffsets(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with count and limit and offset",
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
					"name": "The Associate",
					"rating": 4.2,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-f62bb529-3508-529d-8098-f97f9b67824c"
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
			testUtils.Request{
				Request: `query {
					Author {
						name
						_count(published: {offset: 1, limit: 2})
						published(limit: 2) {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":   "John Grisham",
							"_count": 2,
							"published": []map[string]any{
								{
									"name": "The Associate",
								},
								{
									"name": "Painted House",
								},
							},
						},
						{
							"name":   "Cornelia Funke",
							"_count": 0,
							"published": []map[string]any{
								{
									"name": "Theif Lord",
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

func TestQueryOneToManyWithCountWithLimitWithOffset(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with count with limit with offset",
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
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-f62bb529-3508-529d-8098-f97f9b67824c"
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
			testUtils.Request{
				Request: `query {
					Author {
						name
						_count(published: {offset: 1, limit: 1})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":   "John Grisham",
							"_count": 1,
						},
						{
							"name":   "Cornelia Funke",
							"_count": 0,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
