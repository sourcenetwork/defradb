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

// This test is for documentation reasons only. This is not
// desired behaviour (should just return empty).
// func TestQueryOneToManyWithUnknownCidAndDocID(t *testing.T) {
// 	test := testUtils.RequestTestCase{
// 		Request: `query {
// 					Book (
// 							cid: "bafybeicgwjdyqyuntdop5ytpsfrqg5a4t2r25pfv6prfppl5ta5k5altca",
// 							docID: "bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25"
// 						) {
// 						name
// 						author {
// 							name
// 						}
// 					}
// 				}`,
// 		Docs: map[int][]string{
// 			//books
// 			0: { // bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25
// 				`{
// 					"name": "Painted House",
// 					"rating": 4.9,
// 					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
// 				}`,
// 			},
// 			//authors
// 			1: { // bae-9d52c335-c8e3-5782-8daa-e359c106e0ab
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
		Actions: []any{
			&action.AddSchema{
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
				// bae-f2fa23d1-e9da-5e35-9446-90a80db3c7b7
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-9d52c335-c8e3-5782-8daa-e359c106e0ab
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book (
							cid: "bafyreicmgsatxmz7cksel6m3kws6p6xpr4l7u7hyticxwre6febzgmcupa"
							docID: "bae-f2fa23d1-e9da-5e35-9446-90a80db3c7b7"
						) {
						name
						author {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"name": "Painted House",
							"author": map[string]any{
								"name": "John Grisham",
							},
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
		Actions: []any{
			&action.AddSchema{
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
				// bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-9d52c335-c8e3-5782-8daa-e359c106e0ab
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
							cid: "bafyreicmgsatxmz7cksel6m3kws6p6xpr4l7u7hyticxwre6febzgmcupa",
							docID: "bae-f2fa23d1-e9da-5e35-9446-90a80db3c7b7"
						) {
						name
						author {
							name
							age
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithParentUpdateAndFirstCidAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
				// bae-8627532a-2ed3-50ed-91d5-26f6b9b44c25
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-9d52c335-c8e3-5782-8daa-e359c106e0ab
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
						cid: "bafyreicmgsatxmz7cksel6m3kws6p6xpr4l7u7hyticxwre6febzgmcupa",
						docID: "bae-f2fa23d1-e9da-5e35-9446-90a80db3c7b7"
					) {
						name
						rating
						author {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToManyWithParentUpdateAndLastCidAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
				// bae-f2fa23d1-e9da-5e35-9446-90a80db3c7b7
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-9d52c335-c8e3-5782-8daa-e359c106e0ab
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
						cid: "bafyreifxvqatsma2slodnnxylgbgp75tqpbfuviwz4d2a75y7gwlozevmq",
						docID: "bae-f2fa23d1-e9da-5e35-9446-90a80db3c7b7"
					) {
						name
						rating
						author {
							name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
