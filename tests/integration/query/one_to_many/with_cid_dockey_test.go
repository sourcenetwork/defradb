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

	"github.com/stretchr/testify/assert"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// This test is for documentation reasons only. This is not
// desired behaviour (should just return empty).
func TestQueryOneToManyWithUnknownCidAndDocKey(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "One-to-many relation query from one side with unknown cid and dockey",
		Query: `query {
					book (
							cid: "bafybeihtn2xjbjjqxeqp2uhwhvk3tmjfkaf2qtfqh5w5q3ews7ax2dc75a",
							dockey: "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
						) {
						name
						author {
							name
						}
					}
				}`,
		Docs: map[int][]string{
			//books
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				(`{
				"name": "Painted House",
				"rating": 4.9,
				"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
			}`)},
			//authors
			1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				(`{
				"name": "John Grisham",
				"age": 65,
				"verified": true
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"name": "Painted House",
				"author": map[string]interface{}{
					"name": "John Grisham",
				},
			},
		},
	}

	assert.Panics(t, func() {
		executeTestCase(t, test)
	})
}

func TestQueryOneToManyWithCidAndDocKey(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "One-to-many relation query from one side with  cid and dockey",
		Query: `query {
					book (
							cid: "bafybeigcxmx2mbkmmkujm6vv4eoa57vfg2a22sum4p46empn6fcqkzpdma",
							dockey: "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
						) {
						name
						author {
							name
						}
					}
				}`,
		Docs: map[int][]string{
			//books
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				(`{
				"name": "Painted House",
				"rating": 4.9,
				"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
			}`)},
			//authors
			1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				(`{
				"name": "John Grisham",
				"age": 65,
				"verified": true
			}`)},
		},
		Results: []map[string]interface{}{
			{
				"name": "Painted House",
				"author": map[string]interface{}{
					"name": "John Grisham",
				},
			},
		},
	}

	executeTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (no way to get state of child a time of
// parent creation without explicit child cid, which is also not tied
// to parent state).
func TestQueryOneToManyWithChildUpdateAndFirstCidAndDocKey(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "One-to-many relation query from one side with child update and parent cid and dockey",
		Query: `query {
					book (
							cid: "bafybeigcxmx2mbkmmkujm6vv4eoa57vfg2a22sum4p46empn6fcqkzpdma",
							dockey: "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
						) {
						name
						author {
							name
							age
						}
					}
				}`,
		Docs: map[int][]string{
			//books
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				(`{
				"name": "Painted House",
				"rating": 4.9,
				"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
			}`)},
			//authors
			1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				(`{
				"name": "John Grisham",
				"age": 65,
				"verified": true
			}`)},
		},
		Updates: map[int]map[int][]string{
			1: {
				0: {
					`{
						"age": 22
					}`,
				},
			},
		},
		Results: []map[string]interface{}{
			{
				"name": "Painted House",
				"author": map[string]interface{}{
					"name": "John Grisham",
					"age":  uint64(22),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithParentUpdateAndFirstCidAndDocKey(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "One-to-many relation query from one side with parent update and parent cid and dockey",
		Query: `query {
					book (
							cid: "bafybeigcxmx2mbkmmkujm6vv4eoa57vfg2a22sum4p46empn6fcqkzpdma",
							dockey: "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
						) {
						name
						rating
						author {
							name
						}
					}
				}`,
		Docs: map[int][]string{
			//books
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				(`{
				"name": "Painted House",
				"rating": 4.9,
				"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
			}`)},
			//authors
			1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				(`{
				"name": "John Grisham",
				"age": 65,
				"verified": true
			}`)},
		},
		Updates: map[int]map[int][]string{
			0: {
				0: {
					`{
						"rating": 4.5
					}`,
				},
			},
		},
		Results: []map[string]interface{}{
			{
				"name":   "Painted House",
				"rating": float64(4.9),
				"author": map[string]interface{}{
					"name": "John Grisham",
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithParentUpdateAndLastCidAndDocKey(t *testing.T) {
	test := testUtils.QueryTestCase{
		Description: "One-to-many relation query from one side with parent update and parent cid and dockey",
		Query: `query {
					book (
							cid: "bafybeifc33ql7a5vna3epx55lm2dwyhmq7souodlhnsrfwhz6gfqcb6wje",
							dockey: "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
						) {
						name
						rating
						author {
							name
						}
					}
				}`,
		Docs: map[int][]string{
			//books
			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				(`{
				"name": "Painted House",
				"rating": 4.9,
				"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
			}`)},
			//authors
			1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				(`{
				"name": "John Grisham",
				"age": 65,
				"verified": true
			}`)},
		},
		Updates: map[int]map[int][]string{
			0: {
				0: {
					`{
						"rating": 4.5
					}`,
				},
			},
		},
		Results: []map[string]interface{}{
			{
				"name":   "Painted House",
				"rating": float64(4.5),
				"author": map[string]interface{}{
					"name": "John Grisham",
				},
			},
		},
	}

	executeTestCase(t, test)
}
