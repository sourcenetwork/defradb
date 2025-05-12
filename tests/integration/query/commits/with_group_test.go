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
				Request: ` {
						commits(groupBy: [height]) {
							height
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"height": int64(2),
						},
						{
							"height": int64(1),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithGroupByHeightWithChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by height",
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
				Request: ` {
						commits(groupBy: [height]) {
							height
							_group {
								cid
							}
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"height": int64(2),
							"_group": []map[string]any{
								{
									"cid": "bafyreic5mqzoba47yzm5pugx5b35visawxi2al2tq7p7x2b6yayklwomga",
								},
								{
									"cid": "bafyreido4fwolghako5ogh4jcy6tr3butjicfwubk27uyuimlm366rtdmy",
								},
							},
						},
						{
							"height": int64(1),
							"_group": []map[string]any{
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
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
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
						"name":	"John",
						"age":	21
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
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreiae763hq5srsefplqrehpsuyieuwmbvblgzdma7srss522yciumhu",
							"_group": []map[string]any{
								{
									"height": int64(1),
								},
							},
						},
						{
							"cid": "bafyreiht7yhnnrgbwgyu5toe3exvpkovzrefzr6midu5secnlr546oel3q",
							"_group": []map[string]any{
								{
									"height": int64(1),
								},
							},
						},
						{
							"cid": "bafyreidtdklweht7ainl5rrdeqscr3cwr72sr4lehzrpmmnnbvnvstavnm",
							"_group": []map[string]any{
								{
									"height": int64(1),
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

func TestQueryCommitsWithGroupByDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by document ID",
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
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        1,
				Doc: `{
					"age":	26
				}`,
			},
			testUtils.Request{
				Request: ` {
						commits(groupBy: [docID]) {
							docID
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"docID": "bae-2bb3e007-c40c-5264-8656-45e024cc4776",
						},
						{
							"docID": "bae-dfeea2ca-5e6d-5333-85e8-213a80b508f7",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithGroupByFieldName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by fieldName",
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
				Request: ` {
						commits(groupBy: [fieldName]) {
							fieldName
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"fieldName": "age",
						},
						{
							"fieldName": "name",
						},
						{
							"fieldName": "_C",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithGroupByFieldNameWithChild(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, group by fieldName",
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
				Request: ` {
						commits(groupBy: [fieldName]) {
							fieldName
							_group {
								height
							}
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"fieldName": "age",
							"_group": []map[string]any{
								{
									"height": int64(2),
								},
								{
									"height": int64(1),
								},
							},
						},
						{
							"fieldName": "name",
							"_group": []map[string]any{
								{
									"height": int64(1),
								},
							},
						},
						{
							"fieldName": "_C",
							"_group": []map[string]any{
								{
									"height": int64(2),
								},
								{
									"height": int64(1),
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
