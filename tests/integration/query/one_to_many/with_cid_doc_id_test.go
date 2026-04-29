// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package one_to_many

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// This test is for documentation reasons only. This is not
// desired behaviour (should just return empty).
// func TestQueryOneToManyWithUnknownCidAndDocID(t *testing.T) {
// 	test := &action.Request{TestCase{
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
// 					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				// bae-82bbdc18-aa15-57b8-83af-795a752b3b8f
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				// bae-9d52c335-c8e3-5782-8daa-e359c106e0ab
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.Request{
				Request: `query {
					Book (
							cid: "bafyreial4br7zz2teyhegjcijy2hw6i3oirvxyjxdbrjjnhxmhkphd3l2q"
							docID: "bae-82bbdc18-aa15-57b8-83af-795a752b3b8f"
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				// bae-82bbdc18-aa15-57b8-83af-795a752b3b8f
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				// bae-9d52c335-c8e3-5782-8daa-e359c106e0ab
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.UpdateDoc{
				CollectionID: 1,
				Doc: `{
					"age": 22
				}`,
			},
			&action.Request{
				Request: `query {
					Book (
							cid: "bafyreial4br7zz2teyhegjcijy2hw6i3oirvxyjxdbrjjnhxmhkphd3l2q",
							docID: "bae-82bbdc18-aa15-57b8-83af-795a752b3b8f"
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				// bae-82bbdc18-aa15-57b8-83af-795a752b3b8f
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				// bae-9d52c335-c8e3-5782-8daa-e359c106e0ab
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.UpdateDoc{
				CollectionID: 0,
				Doc: `{
					"rating": 4.5
				}`,
			},
			&action.Request{
				Request: `query {
					Book (
						cid: "bafyreial4br7zz2teyhegjcijy2hw6i3oirvxyjxdbrjjnhxmhkphd3l2q",
						docID: "bae-82bbdc18-aa15-57b8-83af-795a752b3b8f"
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
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				// bae-82bbdc18-aa15-57b8-83af-795a752b3b8f
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"_authorID": "bae-9d52c335-c8e3-5782-8daa-e359c106e0ab"
				}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				// bae-9d52c335-c8e3-5782-8daa-e359c106e0ab
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			&action.UpdateDoc{
				CollectionID: 0,
				Doc: `{
					"rating": 4.5
				}`,
			},
			&action.Request{
				Request: `query {
					Book (
						cid: "bafyreifxbqpzvepc2rseagci6beohmv3qr3knjnfddzk7oqru5su7bdtpi",
						docID: "bae-82bbdc18-aa15-57b8-83af-795a752b3b8f"
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
