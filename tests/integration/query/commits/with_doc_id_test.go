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

func TestQueryCommitsWithUnknownDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with unknown document ID",
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
						commits(docID: "unknown document ID") {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID",
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
						commits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7") {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
						},
						{
							"cid": "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
						},
						{
							"cid": "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, with links",
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
						commits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":   "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
							"links": []map[string]any{},
						},
						{
							"cid":   "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
							"links": []map[string]any{},
						},
						{
							"cid": "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
							"links": []map[string]any{
								{
									"cid":  "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
									"name": "age",
								},
								{
									"cid":  "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
									"name": "name",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, multiple results",
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
						commits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7") {
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
							"cid":    "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
							"height": int64(1),
						},
						{
							"cid":    "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
							"height": int64(1),
						},
						{
							"cid":    "bafyreido4fwolghako5ogh4jcy6tr3butjicfwubk27uyuimlm366rtdmy",
							"height": int64(2),
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

// This test is for documentation reasons only. This is not
// desired behaviour (first results includes link._head, second
// includes link._Name).
func TestQueryCommitsWithDocIDAndUpdateAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID, multiple results and links",
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
						commits(docID: "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreic5mqzoba47yzm5pugx5b35visawxi2al2tq7p7x2b6yayklwomga",
							"links": []map[string]any{
								{
									"cid":  "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
									"name": "_head",
								},
							},
						},
						{
							"cid":   "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
							"links": []map[string]any{},
						},
						{
							"cid":   "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
							"links": []map[string]any{},
						},
						{
							"cid": "bafyreido4fwolghako5ogh4jcy6tr3butjicfwubk27uyuimlm366rtdmy",
							"links": []map[string]any{
								{
									"cid":  "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
									"name": "_head",
								},
								{
									"cid":  "bafyreic5mqzoba47yzm5pugx5b35visawxi2al2tq7p7x2b6yayklwomga",
									"name": "age",
								},
							},
						},
						{
							"cid": "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
							"links": []map[string]any{
								{
									"cid":  "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
									"name": "age",
								},
								{
									"cid":  "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
									"name": "name",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
