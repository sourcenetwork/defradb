// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsWithDepth1(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with depth 1",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits(depth: 1) {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeibcr5lkdvcvtr67rpsnvn57hgrhlg36cnmmf7kywjekjodwxytpi4",
					},
					{
						"cid": "bafybeifw7cu7uweruypv44on2zupjzolyqvyh4ookoeybkztzys67m4hwi",
					},
					{
						"cid": "bafybeige7qoom3bgjitfisxvhbifou6n4tgguan3ihwbkz5mvbumndeiaa",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDepth1WithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with depth 1, and doc updates",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `query {
						commits(depth: 1) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						// "Age" field head
						"cid":    "bafybeibybndrw4dida2m2mlwfl42i56pwoxna6ztansrvya4ikejd63kju",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeifw7cu7uweruypv44on2zupjzolyqvyh4ookoeybkztzys67m4hwi",
						"height": int64(1),
					},
					{
						"cid":    "bafybeieufqlniob4m5abilofa7iewl3mheykvordbhuhi5g4ewszmxnfvi",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDepth2WithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with depth 2, and doc updates",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	23
				}`,
			},
			testUtils.Request{
				Request: `query {
						commits(depth: 2) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						// Composite head
						"cid":    "bafybeifj3dw2wehaabwmrkcmebj3xyyujlp32sycydd3wfjszx3bfxglfu",
						"height": int64(3),
					},
					{
						// Composite head -1
						"cid":    "bafybeibybndrw4dida2m2mlwfl42i56pwoxna6ztansrvya4ikejd63kju",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeifw7cu7uweruypv44on2zupjzolyqvyh4ookoeybkztzys67m4hwi",
						"height": int64(1),
					},
					{
						// "Age" field head
						"cid":    "bafybeieirgdstog2griwuuxgb4c3frgka55yoodjwdznraoieqcxfdijw4",
						"height": int64(3),
					},
					{
						// "Age" field head -1
						"cid":    "bafybeieufqlniob4m5abilofa7iewl3mheykvordbhuhi5g4ewszmxnfvi",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDepth1AndMultipleDocs(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with depth 1",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Fred",
						"age":	25
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits(depth: 1) {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeiasu5mdp6652oux4avwugv6gbd6ciqqsuj2zjv4ypksmiwndgwkeq",
					},
					{
						"cid": "bafybeia7shc4tpafpzblxqjyxmb7fayegsvaol3p2ucujaawig3wtopibu",
					},
					{
						"cid": "bafybeifwn57hy5m5rddplfxdomes34ykck775yvinc522nowspkvawqr6q",
					},
					{
						"cid": "bafybeibcr5lkdvcvtr67rpsnvn57hgrhlg36cnmmf7kywjekjodwxytpi4",
					},
					{
						"cid": "bafybeifw7cu7uweruypv44on2zupjzolyqvyh4ookoeybkztzys67m4hwi",
					},
					{
						"cid": "bafybeige7qoom3bgjitfisxvhbifou6n4tgguan3ihwbkz5mvbumndeiaa",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
