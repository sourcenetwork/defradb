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
// func TestQueryOneToManyWithUnknownCidAndDocID(t *testing.T) {
// 	test := testUtils.RequestTestCase{
// 		Description: "One-to-many relation query from one side with unknown cid and docID",
// 		Request: `query {
// 					Book (
// 							cid: "bafybeicgwjdyqyuntdop5ytpsfrqg5a4t2r25pfv6prfppl5ta5k5altca",
// 							docID: "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
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

func TestQueryOneToManyWithCidAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from one side with cid and docID",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						rating: Float
						author: Author
					}
				
					type Author {
						name: String
						age: Int
						verified: Boolean
						published: [Book]
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book (
							cid: "bafybeia3qbhebdwssoe5udinpbdj4pntb5wjr77ql7ptzq32howbaxz2cu"
							docID: "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d"
						) {
						name
						author {
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Painted House",
						"author": map[string]any{
							"name": "John Grisham",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (no way to get state of child a time of
// parent creation without explicit child cid, which is also not tied
// to parent state).
func TestQueryOneToManyWithChildUpdateAndFirstCidAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from one side with child update and parent cid and docID",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						rating: Float
						author: Author
					}
				
					type Author {
						name: String
						age: Int
						verified: Boolean
						published: [Book]
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 1,
				Doc: `{
					"age": 22
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book (
							cid: "bafybeia3qbhebdwssoe5udinpbdj4pntb5wjr77ql7ptzq32howbaxz2cu",
							docID: "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d"
						) {
						name
						author {
							name
							age
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name": "Painted House",
						"author": map[string]any{
							"name": "John Grisham",
							"age":  int64(22),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithParentUpdateAndFirstCidAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from one side with parent update and parent cid and docID",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						rating: Float
						author: Author
					}
				
					type Author {
						name: String
						age: Int
						verified: Boolean
						published: [Book]
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				Doc: `{
					"rating": 4.5
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book (
						cid: "bafybeia3qbhebdwssoe5udinpbdj4pntb5wjr77ql7ptzq32howbaxz2cu",
						docID: "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d"
					) {
						rating
						author {
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "Painted House",
						"rating": float64(4.9),
						"author": map[string]any{
							"name": "John Grisham",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithParentUpdateAndLastCidAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation query from one side with parent update and parent cid and docID",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						rating: Float
						author: Author
					}
				
					type Author {
						name: String
						age: Int
						verified: Boolean
						published: [Book]
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-fd541c25-229e-5280-b44b-e5c2af3e374d
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-41598f0c-19bc-5da6-813b-e80f14a10df3"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-41598f0c-19bc-5da6-813b-e80f14a10df3
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				Doc: `{
					"rating": 4.5
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book (
						cid: "bafybeibqkdnc63xh5k4frs3x3k7z7p6sw4usjrhxd4iusbjj2uhxfjfjcq",
						docID: "bae-b9b83269-1f28-5c3b-ae75-3fb4c00d559d"
					) {
						rating
						author {
							name
						}
					}
				}`,
				Results: []map[string]any{
					{
						"name":   "Painted House",
						"rating": float64(4.5),
						"author": map[string]any{
							"name": "John Grisham",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
