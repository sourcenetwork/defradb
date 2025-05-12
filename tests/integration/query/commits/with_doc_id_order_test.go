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

func TestQueryCommitsWithDocIDAndOrderHeightDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order height desc",
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
						commits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7", order: {height: DESC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":    "bafyreic5mqzoba47yzm5pugx5b35visawxi2al2tq7p7x2b6yayklwomga",
							"height": int64(2),
						},
						{
							"cid":    "bafyreido4fwolghako5ogh4jcy6tr3butjicfwubk27uyuimlm366rtdmy",
							"height": int64(2),
						},
						{
							"cid":    "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
							"height": int64(1),
						},
						{
							"cid":    "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
							"height": int64(1),
						},
						{
							"cid":    "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
							"height": int64(1),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndOrderHeightAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order height asc",
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
						commits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7", order: {height: ASC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":    "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
							"height": int64(1),
						},
						{
							"cid":    "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
							"height": int64(1),
						},
						{
							"cid":    "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
							"height": int64(1),
						},
						{
							"cid":    "bafyreic5mqzoba47yzm5pugx5b35visawxi2al2tq7p7x2b6yayklwomga",
							"height": int64(2),
						},
						{
							"cid":    "bafyreido4fwolghako5ogh4jcy6tr3butjicfwubk27uyuimlm366rtdmy",
							"height": int64(2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndOrderCidDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order cid desc",
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
						commits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7", order: {cid: DESC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":    "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
							"height": int64(1),
						},
						{
							"cid":    "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
							"height": int64(1),
						},
						{
							"cid":    "bafyreido4fwolghako5ogh4jcy6tr3butjicfwubk27uyuimlm366rtdmy",
							"height": int64(2),
						},
						{
							"cid":    "bafyreic5mqzoba47yzm5pugx5b35visawxi2al2tq7p7x2b6yayklwomga",
							"height": int64(2),
						},
						{
							"cid":    "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
							"height": int64(1),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndOrderCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, order cid asc",
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
						commits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7", order: {cid: ASC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":    "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
							"height": int64(1),
						},
						{
							"cid":    "bafyreic5mqzoba47yzm5pugx5b35visawxi2al2tq7p7x2b6yayklwomga",
							"height": int64(2),
						},
						{
							"cid":    "bafyreido4fwolghako5ogh4jcy6tr3butjicfwubk27uyuimlm366rtdmy",
							"height": int64(2),
						},
						{
							"cid":    "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
							"height": int64(1),
						},
						{
							"cid":    "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
							"height": int64(1),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndOrderAndMultiUpdatesCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, multiple updates with order cid asc",
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
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	24
				}`,
			},
			testUtils.Request{
				Request: `query {
						 commits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7", order: {height: ASC}) {
							 cid
							 height
						 }
					 }`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":    "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
							"height": int64(1),
						},
						{
							"cid":    "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
							"height": int64(1),
						},
						{
							"cid":    "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
							"height": int64(1),
						},
						{
							"cid":    "bafyreic5mqzoba47yzm5pugx5b35visawxi2al2tq7p7x2b6yayklwomga",
							"height": int64(2),
						},
						{
							"cid":    "bafyreido4fwolghako5ogh4jcy6tr3butjicfwubk27uyuimlm366rtdmy",
							"height": int64(2),
						},
						{
							"cid":    "bafyreiaa6r6d6ure3murz63ebfbmwlmtgm5rc5wbqvucufnku2k3vlgsga",
							"height": int64(3),
						},
						{
							"cid":    "bafyreihsykzvfrxzsq6tqdzbebs2dgqfy3n5rxfy5b5zrbjdp6ktzlr56m",
							"height": int64(3),
						},
						{
							"cid":    "bafyreibicj5tx6xtrf3hd52i7kpkhunkbaq636potkkcwqkr2c6bs5smea",
							"height": int64(4),
						},
						{
							"cid":    "bafyreib4h46jgp3k6ykwzlp2uuqkqqhbzemvkt74ivtmnbmpy6cc7ltpum",
							"height": int64(4),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
