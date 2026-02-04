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

func TestOneToManyAscOrderAndFilterOnParentWithAggSumOnSubTypeField(t *testing.T) {
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
					"name": "The Associate",
					"rating": 4.2,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Sooley",
					"rating": 3.2,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Rooster Bar",
					"rating": 4,
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
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Little Kid",
					"age": 6,
					"verified": true
				}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Not a Writer",
					"age": 85,
					"verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Author(order: {age: ASC}, filter: {age: {_gt: 8}}) {
						name
						_sum(published: {field: rating})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "Cornelia Funke",
							"_sum": 4.8,
						},
						{
							"name": "John Grisham",
							"_sum": 4.9 + 4.5 + 4.2 + 3.2 + 4.0,
						},
						{
							"name": "Not a Writer",
							"_sum": 0.0,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestOneToManyDescOrderAndFilterOnParentWithAggSumOnSubTypeField(t *testing.T) {
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
					"name": "The Associate",
					"rating": 4.2,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Sooley",
					"rating": 3.2,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Rooster Bar",
					"rating": 4,
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
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Little Kid",
					"age": 6,
					"verified": true
				}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Not a Writer",
					"age": 85,
					"verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Author(order: {age: DESC}, filter: {age: {_gt: 8}}) {
						name
						_sum(published: {field: rating})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "Not a Writer",
							"_sum": 0.0,
						},
						{
							"name": "John Grisham",
							"_sum": 4.9 + 4.5 + 4.2 + 3.2 + 4.0,
						},
						{
							"name": "Cornelia Funke",
							"_sum": 4.8,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestOnetoManySumBySubTypeFieldAndSumBySybTypeFieldWithDescOrderingOnFieldWithLimit(t *testing.T) {
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
					"name": "The Associate",
					"rating": 4.2,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Sooley",
					"rating": 3.2,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Rooster Bar",
					"rating": 4,
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
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Little Kid",
					"age": 6,
					"verified": true
				}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Not a Writer",
					"age": 85,
					"verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Author {
						name
						sum1: _sum(published: {field: rating})
						sum2: _sum(published: {field: rating, limit: 2, order: {rating: DESC}})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "Little Kid",
							"sum1": 0.0,
							"sum2": 0.0,
						},
						{
							"name": "Cornelia Funke",
							"sum1": 4.8,
							"sum2": 4.8,
						},
						{
							"name": "Not a Writer",
							"sum1": 0.0,
							"sum2": 0.0,
						},
						{
							"name": "John Grisham",
							"sum1": 4.9 + 4.5 + 4.2 + 3.2 + 4.0,
							"sum2": 4.9 + 4.5,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestOnetoManySumBySubTypeFieldAndSumBySybTypeFieldWithAscOrderingOnFieldWithLimit(t *testing.T) {
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
					"name": "The Associate",
					"rating": 4.2,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Sooley",
					"rating": 3.2,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Rooster Bar",
					"rating": 4,
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
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Little Kid",
					"age": 6,
					"verified": true
				}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Not a Writer",
					"age": 85,
					"verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Author {
						name
						sum1: _sum(published: {field: rating})
						sum2: _sum(published: {field: rating, limit: 2, order: {rating: ASC}})
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"name": "Little Kid",
							"sum1": 0.0,
							"sum2": 0.0,
						},
						{
							"name": "Cornelia Funke",
							"sum1": 4.8,
							"sum2": 4.8,
						},
						{
							"name": "Not a Writer",
							"sum1": 0.0,
							"sum2": 0.0,
						},
						{
							"name": "John Grisham",
							"sum1": 4.9 + 4.5 + 4.2 + 3.2 + 4.0,
							"sum2": 4.0 + 3.2,
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestOneToManyLimitAscOrderSumOfSubTypeAndLimitAscOrderFieldsOfSubtype(t *testing.T) {
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
					"name": "The Associate",
					"rating": 4.2,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Sooley",
					"rating": 3.2,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Rooster Bar",
					"rating": 4,
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
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Little Kid",
					"age": 6,
					"verified": true
				}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Not a Writer",
					"age": 85,
					"verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Author {
						LimitOrderSum: _sum(published: {field: rating, limit: 2, order: {rating: ASC}})
						LimitOrderFields: published(order: {rating: ASC}, limit: 2) {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"LimitOrderSum":    0.0,
							"LimitOrderFields": []map[string]any{},
						},
						{
							"LimitOrderSum": 4.8,
							"LimitOrderFields": []map[string]any{
								{
									"name": "Theif Lord",
								},
							},
						},
						{
							"LimitOrderSum":    0.0,
							"LimitOrderFields": []map[string]any{},
						},
						{
							"LimitOrderSum": 3.2 + 4.0,
							"LimitOrderFields": []map[string]any{
								{
									"name": "Sooley",
								},
								{
									"name": "The Rooster Bar",
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

func TestOneToManyLimitDescOrderSumOfSubTypeAndLimitAscOrderFieldsOfSubtype(t *testing.T) {
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
					"name": "The Associate",
					"rating": 4.2,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Sooley",
					"rating": 3.2,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "The Rooster Bar",
					"rating": 4,
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
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Little Kid",
					"age": 6,
					"verified": true
				}`,
			},
			&action.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Not a Writer",
					"age": 85,
					"verified": false
				}`,
			},
			&action.Request{
				Request: `query {
					Author {
						LimitOrderSum: _sum(published: {field: rating, limit: 2, order: {rating: DESC}})
						LimitOrderFields: published(order: {rating: DESC}, limit: 2) {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Author": []map[string]any{
						{
							"LimitOrderSum":    0.0,
							"LimitOrderFields": []map[string]any{},
						},
						{
							"LimitOrderSum": 4.8,
							"LimitOrderFields": []map[string]any{
								{
									"name": "Theif Lord",
								},
							},
						},
						{
							"LimitOrderSum":    0.0,
							"LimitOrderFields": []map[string]any{},
						},
						{
							"LimitOrderSum": 4.9 + 4.5,
							"LimitOrderFields": []map[string]any{
								{
									"name": "Painted House",
								},
								{
									"name": "A Time for Mercy",
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
