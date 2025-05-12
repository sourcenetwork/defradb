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
// 		Description: "One-to-many relation query from one side with unknown cid and docID",
// 		Request: `query {
// 					Book (
// 							cid: "bafybeicgwjdyqyuntdop5ytpsfrqg5a4t2r25pfv6prfppl5ta5k5altca",
// 							docID: "bae-818aecea-02f9-5064-9e17-c8b7cc20e63f"
// 						) {
// 						name
// 						author {
// 							name
// 						}
// 					}
// 				}`,
// 		Docs: map[int][]string{
// 			//books
// 			0: { // bae-818aecea-02f9-5064-9e17-c8b7cc20e63f
// 				`{
// 					"name": "Painted House",
// 					"rating": 4.9,
// 					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
// 				}`,
// 			},
// 			//authors
// 			1: { // bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80
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
				// bae-818aecea-02f9-5064-9e17-c8b7cc20e63f
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80
				Doc: `{
					"name": "John Grisham",
					"age": 65,
					"verified": true
				}`,
			},
			testUtils.Request{
				Request: `query {
					Book (
							cid: "bafyreibswgqxe2qfpmj5tusvb6y6y5e36zftioiruxa5fznhks22kmrtle"
							docID: "bae-54426e27-e18b-5b9e-9bbd-edfa36f6bbc4"
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
		Description: "One-to-many relation query from one side with child update and parent cid and docID",
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
				// bae-818aecea-02f9-5064-9e17-c8b7cc20e63f
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80
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
							cid: "bafyreibswgqxe2qfpmj5tusvb6y6y5e36zftioiruxa5fznhks22kmrtle",
							docID: "bae-54426e27-e18b-5b9e-9bbd-edfa36f6bbc4"
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
		Description: "One-to-many relation query from one side with parent update and parent cid and docID",
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
				// bae-818aecea-02f9-5064-9e17-c8b7cc20e63f
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80
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
						cid: "bafyreibswgqxe2qfpmj5tusvb6y6y5e36zftioiruxa5fznhks22kmrtle",
						docID: "bae-54426e27-e18b-5b9e-9bbd-edfa36f6bbc4"
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
		Description: "One-to-many relation query from one side with parent update and parent cid and docID",
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
				// bae-54426e27-e18b-5b9e-9bbd-edfa36f6bbc4
				Doc: `{
					"name": "Painted House",
					"rating": 4.9,
					"author_id": "bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80"
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				// bae-c0ecb296-4f8b-5037-a0e7-f10d8d5d5b80
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
						cid: "bafyreicpmvspecioboq7djbccmn5du7msfbjk2kcdmu6dxzzp5uzlybh64",
						docID: "bae-54426e27-e18b-5b9e-9bbd-edfa36f6bbc4"
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
