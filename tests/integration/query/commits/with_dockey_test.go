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

func TestQueryCommitsWithUnknownDockey(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with unknown dockey",
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
				Request: `query {
						commits(dockey: "unknown dockey") {
							cid
						}
					}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithDockey(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey",
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
				Request: `query {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							cid
						}
					}`,
				Results: []map[string]any{
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
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithDockeyAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, with links",
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
				Request: `query {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid":   "bafybeih7athlviqvv255ujgk6mikxcwxp46n7yyaekxldrgsvgrkrgajtq",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeifez5xzi6x4w6cahxzlhsjbzbarzgtoixfioqr5l4uh2akkgl6hhq",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeig32cpgkhikjerzp33wvbilrcbhm7xqpjeba5ptpgg2s5ivqlndba",
						"links": []map[string]any{
							{
								"cid":  "bafybeih7athlviqvv255ujgk6mikxcwxp46n7yyaekxldrgsvgrkrgajtq",
								"name": "Age",
							},
							{
								"cid":  "bafybeifez5xzi6x4w6cahxzlhsjbzbarzgtoixfioqr5l4uh2akkgl6hhq",
								"name": "Name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithDockeyAndUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple results",
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
				Request: `query {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeigtudzrntyslfdtukiobzen76iuhkmyo3x4eutx3igwzwhjekp2oi",
						"height": int64(2),
					},
					{
						"cid":    "bafybeih7athlviqvv255ujgk6mikxcwxp46n7yyaekxldrgsvgrkrgajtq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeifez5xzi6x4w6cahxzlhsjbzbarzgtoixfioqr5l4uh2akkgl6hhq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeibwql353y3bpl24f25t53gl6zxcsp56uvmdrqrvweh6kv63un4pca",
						"height": int64(2),
					},
					{
						"cid":    "bafybeig32cpgkhikjerzp33wvbilrcbhm7xqpjeba5ptpgg2s5ivqlndba",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (first results includes link._head, second
// includes link._Name).
func TestQueryCommitsWithDockeyAndUpdateAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple results and links",
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
				Request: `query {
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeigtudzrntyslfdtukiobzen76iuhkmyo3x4eutx3igwzwhjekp2oi",
						"links": []map[string]any{
							{
								"cid":  "bafybeih7athlviqvv255ujgk6mikxcwxp46n7yyaekxldrgsvgrkrgajtq",
								"name": "_head",
							},
						},
					},
					{
						"cid":   "bafybeih7athlviqvv255ujgk6mikxcwxp46n7yyaekxldrgsvgrkrgajtq",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeifez5xzi6x4w6cahxzlhsjbzbarzgtoixfioqr5l4uh2akkgl6hhq",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeibwql353y3bpl24f25t53gl6zxcsp56uvmdrqrvweh6kv63un4pca",
						"links": []map[string]any{
							{
								"cid":  "bafybeigtudzrntyslfdtukiobzen76iuhkmyo3x4eutx3igwzwhjekp2oi",
								"name": "Age",
							},
							{
								"cid":  "bafybeig32cpgkhikjerzp33wvbilrcbhm7xqpjeba5ptpgg2s5ivqlndba",
								"name": "_head",
							},
						},
					},
					{
						"cid": "bafybeig32cpgkhikjerzp33wvbilrcbhm7xqpjeba5ptpgg2s5ivqlndba",
						"links": []map[string]any{
							{
								"cid":  "bafybeih7athlviqvv255ujgk6mikxcwxp46n7yyaekxldrgsvgrkrgajtq",
								"name": "Age",
							},
							{
								"cid":  "bafybeifez5xzi6x4w6cahxzlhsjbzbarzgtoixfioqr5l4uh2akkgl6hhq",
								"name": "Name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}
