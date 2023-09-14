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

// This test is for documentation reasons only. This is not
// desired behaviour (should just return empty).
// func TestQueryOneToManyWithUnknownCidAndDocKey(t *testing.T) {
// 	test := testUtils.RequestTestCase{
// 		Description: "One-to-many relation query from one side with unknown cid and dockey",
// 		Request: `query {
// 					Book (
// 							cid: "bafybeicgwjdyqyuntdop5ytpsfrqg5a4t2r25pfv6prfppl5ta5k5altca",
// 							dockey: "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
// 						) {
// 						name
// 						author {
// 							name
// 						}
// 					}
// 				}`,
// 		Docs: map[int][]string{
// 			//books
// 			0: { // bae-fd541c25-229e-5280-b44b-e5c2af3e374d
// 				`{
// 					"name": "Painted House",
// 					"rating": 4.9,
// 					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
// 				}`,
// 			},
// 			//authors
// 			1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
// 				`{
// 					"name": "John Grisham",
// 					"age": 65,
// 					"verified": true
// 				}`,
// 			},
// 		},
// 		Results: []map[string]any{
// 			{
// 				"name": "Painted House",
// 				"author": map[string]any{
// 					"name": "John Grisham",
// 				},
// 			},
// 		},
// 	}

// 	testUtils.AssertPanic(t, func() { executeTestCase(t, test) })
// }

func TestQueryOneToManyWithCidAndDocKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from one side with  cid and dockey",
		Request: `query {
					Book (
							cid: "bafybeie3srbs3vyirntnaubjwn7i3cltht3mfbvpiiahxvkw5yvmte2fne"
							dockey: "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d"
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
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},
			//authors
			1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "Painted House",
				"author": map[string]any{
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
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from one side with child update and parent cid and dockey",
		Request: `query {
					Book (
							cid: "bafybeie3srbs3vyirntnaubjwn7i3cltht3mfbvpiiahxvkw5yvmte2fne",
							dockey: "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d"
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
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},
			//authors
			1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
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
		Results: []map[string]any{
			{
				"name": "Painted House",
				"author": map[string]any{
					"name": "John Grisham",
					"age":  uint64(22),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithParentUpdateAndFirstCidAndDocKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from one side with parent update and parent cid and dockey",
		Request: `query {
					Book (
							cid: "bafybeie3srbs3vyirntnaubjwn7i3cltht3mfbvpiiahxvkw5yvmte2fne",
							dockey: "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d"
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
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},
			//authors
			1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
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
		Results: []map[string]any{
			{
				"name":   "Painted House",
				"rating": float64(4.9),
				"author": map[string]any{
					"name": "John Grisham",
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryOneToManyWithParentUpdateAndLastCidAndDocKey(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "One-to-many relation query from one side with parent update and parent cid and dockey",
		Request: `query {
					Book (
							cid: "bafybeiavnr6gu2ccfm2ygc2m2nsqbhnoorhjf2p6f2rq4rkggjz55je5ym",
							dockey: "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d"
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
				`{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},
			//authors
			1: { // bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				`{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
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
		Results: []map[string]any{
			{
				"name":   "Painted House",
				"rating": float64(4.5),
				"author": map[string]any{
					"name": "John Grisham",
				},
			},
		},
	}

	executeTestCase(t, test)
}
