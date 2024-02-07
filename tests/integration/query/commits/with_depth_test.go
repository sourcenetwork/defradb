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
						"cid": "bafybeieikx6l2xead2dzsa5wwy5irxced2eddyq23jkp4csf5igoob7diq",
					},
					{
						"cid": "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
					},
					{
						"cid": "bafybeiggrv6gyhld2dbkspaxsenjejfhnk52pm4mlpyz2q6x4dlnaff2mu",
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
						"cid":    "bafybeiddpjl27ulw2yo4ohup6gr2wob3pwagqw2rbeaxxodv4ljelnu7ve",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiekajrgheumrgamrc4mprmm66ulp2qr75sviowfcriuviggokydbm",
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
						"cid":    "bafybeibfip3j6fr755tjjhlmuqxqywlmgxbalgbnhggq3xj3xvtwe6f6jy",
						"height": int64(3),
					},
					{
						// Composite head -1
						"cid":    "bafybeiddpjl27ulw2yo4ohup6gr2wob3pwagqw2rbeaxxodv4ljelnu7ve",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
						"height": int64(1),
					},
					{
						// "Age" field head
						"cid":    "bafybeigomkxadtuj4vfkb7ix55d2qhnzh24wnxv4gqvbo2s5hdtyf2y7im",
						"height": int64(3),
					},
					{
						// "Age" field head -1
						"cid":    "bafybeiekajrgheumrgamrc4mprmm66ulp2qr75sviowfcriuviggokydbm",
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
						"cid": "bafybeihf2bo3amfoojd3ihliafgtka4jxava5i6rmujktvzon2x4iuzzna",
					},
					{
						"cid": "bafybeic4lebzixuxdsgzqs4nvy6ont7sheognehdbmpr27ha2cvror6v4m",
					},
					{
						"cid": "bafybeietgsruk6dlepqsre27fripbsjnwawpmygnwdxsmpt5vb6q4mkdxi",
					},
					{
						"cid": "bafybeieikx6l2xead2dzsa5wwy5irxced2eddyq23jkp4csf5igoob7diq",
					},
					{
						"cid": "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
					},
					{
						"cid": "bafybeiggrv6gyhld2dbkspaxsenjejfhnk52pm4mlpyz2q6x4dlnaff2mu",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
