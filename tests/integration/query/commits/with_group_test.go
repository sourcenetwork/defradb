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

func TestQueryCommitsWithGroupBy(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by height",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Age":	22
				}`,
			},
			testUtils.Request{
				Request: ` {
						commits(groupBy: [height]) {
							height
						}
					}`,
				Results: []map[string]any{
					{
						"height": int64(2),
					},
					{
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithGroupByHeightWithChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by height",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Age":	22
				}`,
			},
			testUtils.Request{
				Request: ` {
						commits(groupBy: [height]) {
							height
							_group {
								cid
							}
						}
					}`,
				Results: []map[string]any{
					{
						"height": int64(2),
						"_group": []map[string]any{
							{
								"cid": "bafybeigtudzrntyslfdtukiobzen76iuhkmyo3x4eutx3igwzwhjekp2oi",
							},
							{
								"cid": "bafybeibwql353y3bpl24f25t53gl6zxcsp56uvmdrqrvweh6kv63un4pca",
							},
						},
					},
					{
						"height": int64(1),
						"_group": []map[string]any{
							{
								"cid": "bafybeih7athlviqvv255ujgk6mikxcwxp46n7yyaekxldrgsvgrkrgajtq",
							},
							{
								"cid": "bafybeifez5xzi6x4w6cahxzlhsjbzbarzgtoixfioqr5l4uh2akkgl6hhq",
							},
							{
								"cid": "bafybeig32cpgkhikjerzp33wvbilrcbhm7xqpjeba5ptpgg2s5ivqlndba",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

// This is an odd test, but we need to make sure it works
func TestQueryCommitsWithGroupByCidWithChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by cid",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.Request{
				Request: ` {
						commits(groupBy: [cid]) {
							cid
							_group {
								height
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeih7athlviqvv255ujgk6mikxcwxp46n7yyaekxldrgsvgrkrgajtq",
						"_group": []map[string]any{
							{
								"height": int64(1),
							},
						},
					},
					{
						"cid": "bafybeifez5xzi6x4w6cahxzlhsjbzbarzgtoixfioqr5l4uh2akkgl6hhq",
						"_group": []map[string]any{
							{
								"height": int64(1),
							},
						},
					},
					{
						"cid": "bafybeig32cpgkhikjerzp33wvbilrcbhm7xqpjeba5ptpgg2s5ivqlndba",
						"_group": []map[string]any{
							{
								"height": int64(1),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithGroupByDocKey(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by dockey",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"Fred",
						"Age":	25
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Age":	22
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        1,
				Doc: `{
					"Age":	26
				}`,
			},
			testUtils.Request{
				Request: ` {
						commits(groupBy: [dockey]) {
							dockey
						}
					}`,
				Results: []map[string]any{
					{
						"dockey": "bae-52b9170d-b77a-5887-b877-cbdbb99b009f",
					},
					{
						"dockey": "bae-b2103437-f5bd-52b6-99b1-5970412c5201",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}
