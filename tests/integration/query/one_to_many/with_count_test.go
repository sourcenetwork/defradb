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

func TestQueryOneToManyWithCount(t *testing.T) {
	tests := []testUtils.TestCase{
		{
			Description: "One-to-many relation query from many side with count, no child records",
			Actions: []any{
				testUtils.CreateDoc{
					CollectionID: 1,
					Doc: `{
						"name": "John Grisham",
						"age": 65,
						"verified": true
					}`,
				},
				testUtils.Request{
					Request: `query {
						Author {
							name
							_count(published: {})
						}
					}`,
					Results: map[string]any{
						"Author": []map[string]any{
							{
								"name":   "John Grisham",
								"_count": 0,
							},
						},
					},
				},
			},
		},
		{
			Description: "One-to-many relation query from many side with count",
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
						"name": "Theif Lord",
						"rating": 4.8,
						"author_id": "bae-72e8c691-9f20-55e7-9228-8af1cf54cace"
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
						}
					}`,
					Results: map[string]any{
						"Author": []map[string]any{
							{
								"name":   "Cornelia Funke",
								"_count": 1,
							},
							{
								"name":   "John Grisham",
								"_count": 2,
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

func TestQueryOneToMany_WithCountAliasFilter_ShouldMatchAll(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with count alias",
		Actions: []any{
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
					"name":      "Theif Lord",
					"rating":    4.8,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {_alias: {publishedCount: {_gt: 0}}}) {
						name
						publishedCount: _count(published: {})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":           "Cornelia Funke",
							"publishedCount": 1,
						},
						{
							"name":           "John Grisham",
							"publishedCount": 2,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToMany_WithCountAliasFilter_ShouldMatchOne(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from many side with count alias",
		Actions: []any{
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
					"name":      "Theif Lord",
					"rating":    4.8,
					"author_id": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.Request{
				Request: `query {
					Author(filter: {_alias: {publishedCount: {_gt: 1}}}) {
						name
						publishedCount: _count(published: {})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":           "John Grisham",
							"publishedCount": 2,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
