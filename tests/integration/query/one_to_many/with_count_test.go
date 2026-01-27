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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToMany_WithCount_NothingToCount(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
						"name": "John Grisham",
						"age": 65,
						"verified": true
					}`,
			},
			&action.Request{
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
	}

	executeTestCase(t, test)
}

func TestQueryOneToMany_WithCount_ShouldMatchAll(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Painted House",
						"rating": 4.9,
						"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
					}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name": "A Time for Mercy",
						"rating": 4.5,
						"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
					}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name": "Theif Lord",
						"rating": 4.8,
						"_authorID": "bae-3d5a3204-4e55-5236-992a-ce27da27902b"
					}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
						"name": "John Grisham",
						"age": 65,
						"verified": true
					}`,
			},
			&action.CreateDoc{
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
							_count(published: {})
						}
					}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name":   "John Grisham",
							"_count": 2,
						},
						{
							"name":   "Cornelia Funke",
							"_count": 1,
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToMany_WithCountAliasFilter_ShouldMatchAll(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"_authorID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.Request{
				Request: `query {
					Author(filter: {_alias: {publishedCount: {_gt: 0}}}) {
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
						{
							"name":           "Cornelia Funke",
							"publishedCount": 1,
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToMany_WithCountAliasFilter_ShouldMatchOne(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Painted House",
					"rating":    4.9,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "A Time for Mercy",
					"rating":    4.5,
					"_authorID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":      "Theif Lord",
					"rating":    4.8,
					"_authorID": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.Request{
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
