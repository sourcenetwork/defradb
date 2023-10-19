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

func TestQueryOneToOneWithChildBooleanOrderDescending(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-one relation query with simple descending order by sub type",
		Request: `query {
			Book(order: {author: {verified: DESC}}) {
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
			0: {
				// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
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
			1: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
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
					"age":  int64(65),
				},
			},
			{
				"name":   "Theif Lord",
				"rating": 4.8,
				"author": map[string]any{
					"name": "Cornelia Funke",
					"age":  int64(62),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOneWithChildBooleanOrderAscending(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-one relation query with simple ascending order by sub type",
		Request: `query {
			Book(order: {author: {verified: ASC}}) {
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
			0: {
				// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
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
			1: {
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
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
				"name":   "Theif Lord",
				"rating": 4.8,
				"author": map[string]any{
					"name": "Cornelia Funke",
					"age":  int64(62),
				},
			},
			{
				"name":   "Painted House",
				"rating": 4.9,
				"author": map[string]any{
					"name": "John Grisham",
					"age":  int64(65),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOneWithChildIntOrderDescendingWithNoSubTypeFieldsSelected(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Relation query with descending order by sub-type's int field, but only parent fields are selected.",
		Request: `query {
			Book(order: {author: {age: DESC}}) {
				name
				rating
			}
		}`,
		Docs: map[int][]string{
			//books
			0: {
				// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
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
			1: {
				// "bae-3bfe0092-e31f-5ebe-a3ba-fa18fac448a6"
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
				// "bae-08519989-280d-5a4d-90b2-915ea06df3c4"
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
			},
			{
				"name":   "Theif Lord",
				"rating": 4.8,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToOneWithChildIntOrderAscendingWithNoSubTypeFieldsSelected(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Relation query with ascending order by sub-type's int field, but only parent fields are selected.",
		Request: `query {
			Book(order: {author: {age: ASC}}) {
				name
				rating
			}
		}`,
		Docs: map[int][]string{
			//books
			0: {
				// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
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
			1: {
				// "bae-3bfe0092-e31f-5ebe-a3ba-fa18fac448a6"
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true,
					"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
				}`,
				// "bae-08519989-280d-5a4d-90b2-915ea06df3c4"
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
				"name":   "Theif Lord",
				"rating": 4.8,
			},
			{
				"name":   "Painted House",
				"rating": 4.9,
			},
		},
	}

	executeTestCase(t, test)
}
