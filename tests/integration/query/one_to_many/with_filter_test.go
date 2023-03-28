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

func TestQueryOneToManyWithNumericGreaterThanFilterOnParent(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from the many side, simple filter",
		Request: `query {
			Author(filter: {age: {_gt: 63}}) {
				name
				age
				published {
					name
					rating
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
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
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
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "John Grisham",
				"age":  uint64(65),
				"published": []map[string]any{
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
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithNumericGreaterThanChildFilterOnParentWithUnrenderedChild(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from the many side, simple filter",
		Request: `query {
			Author(filter: {published: {rating: {_gt: 4.8}}}) {
				name
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
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
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
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "John Grisham",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithNumericGreaterThanFilterOnParentAndChild(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from the many side, simple filter on root and sub type",
		Request: `query {
			Author(filter: {age: {_gt: 63}}) {
				name
				age
				published(filter: {rating: {_gt: 4.6}}) {
					name
					rating
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
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
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
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "John Grisham",
				"age":  uint64(65),
				"published": []map[string]any{
					{
						"name":   "Painted House",
						"rating": 4.9,
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithMultipleAliasedFilteredChildren(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from the many side, simple filter on root and sub type",
		Request: `query {
			Author {
				name
				age
				p1: published(filter: {rating: {_gt: 4.6}}) {
					name
					rating
				}
				p2: published(filter: {rating: {_lt: 4.6}}) {
					name
					rating
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
					"name": "Theif Lord",
					"rating": 4.8,
					"author_id": "bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04"
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
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "John Grisham",
				"age":  uint64(65),
				"p1": []map[string]any{
					{
						"name":   "Painted House",
						"rating": 4.9,
					},
				},
				"p2": []map[string]any{
					{
						"name":   "A Time for Mercy",
						"rating": 4.5,
					},
				},
			},
			{
				"name": "Cornelia Funke",
				"age":  uint64(62),
				"p1": []map[string]any{
					{
						"name":   "Theif Lord",
						"rating": 4.8,
					},
				},
				"p2": []map[string]any{},
			},
		},
	}

	executeTestCase(t, test)
}
