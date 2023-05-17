// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOneWithNumericFilterOnParent(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-one relation query with simple filter on sub type",
		Request: `query {
					Book {
						name
						rating
						author(filter: {age: {_eq: 65}}) {
							name
							age
						}
					}
				}`,
		Docs: map[int][]string{
			//books
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				`{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			//authors
			1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name":   "Painted House",
				"rating": 4.9,
				"author": map[string]any{
					"name": "John Grisham",
					"age":  uint64(65),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOneWithStringFilterOnChild(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-one relation query with simple filter on parent",
		Request: `query {
					Book(filter: {name: {_eq: "Painted House"}}) {
						name
						rating
						author {
							name
							age
						}
					}
				}`,
		Docs: map[int][]string{
			//books
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				`{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			//authors
			1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name":   "Painted House",
				"rating": 4.9,
				"author": map[string]any{
					"name": "John Grisham",
					"age":  uint64(65),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOneWithBooleanFilterOnChild(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-one relation query with simple sub filter on child",
		Request: `query {
					Book(filter: {author: {verified: {_eq: true}}}) {
						name
						rating
						author {
							name
							age
						}
					}
				}`,
		Docs: map[int][]string{
			//books
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				`{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			//authors
			1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name":   "Painted House",
				"rating": 4.9,
				"author": map[string]any{
					"name": "John Grisham",
					"age":  uint64(65),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOneWithFilterThroughChildBackToParent(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-one relation query with filter on parent referencing parent through child",
		Request: `query {
					Book(filter: {author: {published: {rating: {_eq: 4.9}}}}) {
						name
						rating
						author {
							name
							age
						}
					}
				}`,
		Docs: map[int][]string{
			//books
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				`{
					"name": "Painted House",
					"rating": 4.9
				}`,
				// bae-d432bdfb-787d-5a1c-ac29-dc025ab80095
				`{
					"name": "Theif Lord",
					"rating": 4.8
				}`,
			},
			//authors
			1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
				// bae-b769708d-f552-5c3d-a402-ccfd7ac7fb04
				`{
					"name": "Cornelia Funke",
					"age": 62,
					"verified": false,
					"published_id": "bae-d432bdfb-787d-5a1c-ac29-dc025ab80095"
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name":   "Painted House",
				"rating": 4.9,
				"author": map[string]any{
					"name": "John Grisham",
					"age":  uint64(65),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOneWithBooleanFilterOnChildWithNoSubTypeSelection(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-one relation with simple sub filter on child, but not child selections",
		Request: `query {
					Book(filter: {author: {verified: {_eq: true}}}) {
						name
						rating
					}
				}`,
		Docs: map[int][]string{
			//books
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				`{
					"name": "Painted House",
					"rating": 4.9
				}`,
			},
			//authors
			1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name":   "Painted House",
				"rating": 4.9,
			},
		},
	}

	executeTestCase(t, test)
}
