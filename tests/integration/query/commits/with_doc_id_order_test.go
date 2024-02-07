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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiddpjl27ulw2yo4ohup6gr2wob3pwagqw2rbeaxxodv4ljelnu7ve",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiekajrgheumrgamrc4mprmm66ulp2qr75sviowfcriuviggokydbm",
						"height": int64(2),
					},
					{
						"cid":    "bafybeieikx6l2xead2dzsa5wwy5irxced2eddyq23jkp4csf5igoob7diq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiggrv6gyhld2dbkspaxsenjejfhnk52pm4mlpyz2q6x4dlnaff2mu",
						"height": int64(1),
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeieikx6l2xead2dzsa5wwy5irxced2eddyq23jkp4csf5igoob7diq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiggrv6gyhld2dbkspaxsenjejfhnk52pm4mlpyz2q6x4dlnaff2mu",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiddpjl27ulw2yo4ohup6gr2wob3pwagqw2rbeaxxodv4ljelnu7ve",
						"height": int64(2),
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiggrv6gyhld2dbkspaxsenjejfhnk52pm4mlpyz2q6x4dlnaff2mu",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiekajrgheumrgamrc4mprmm66ulp2qr75sviowfcriuviggokydbm",
						"height": int64(2),
					},
					{
						"cid":    "bafybeieikx6l2xead2dzsa5wwy5irxced2eddyq23jkp4csf5igoob7diq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiddpjl27ulw2yo4ohup6gr2wob3pwagqw2rbeaxxodv4ljelnu7ve",
						"height": int64(2),
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiddpjl27ulw2yo4ohup6gr2wob3pwagqw2rbeaxxodv4ljelnu7ve",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
						"height": int64(1),
					},
					{
						"cid":    "bafybeieikx6l2xead2dzsa5wwy5irxced2eddyq23jkp4csf5igoob7diq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiekajrgheumrgamrc4mprmm66ulp2qr75sviowfcriuviggokydbm",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiggrv6gyhld2dbkspaxsenjejfhnk52pm4mlpyz2q6x4dlnaff2mu",
						"height": int64(1),
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
						 commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							 cid
							 height
						 }
					 }`,
				Results: []map[string]any{
					{
						"cid":    "bafybeieikx6l2xead2dzsa5wwy5irxced2eddyq23jkp4csf5igoob7diq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiggrv6gyhld2dbkspaxsenjejfhnk52pm4mlpyz2q6x4dlnaff2mu",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiddpjl27ulw2yo4ohup6gr2wob3pwagqw2rbeaxxodv4ljelnu7ve",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiekajrgheumrgamrc4mprmm66ulp2qr75sviowfcriuviggokydbm",
						"height": int64(2),
					},
					{
						"cid":    "bafybeibfip3j6fr755tjjhlmuqxqywlmgxbalgbnhggq3xj3xvtwe6f6jy",
						"height": int64(3),
					},
					{
						"cid":    "bafybeigomkxadtuj4vfkb7ix55d2qhnzh24wnxv4gqvbo2s5hdtyf2y7im",
						"height": int64(3),
					},
					{
						"cid":    "bafybeigrdrat2xyrfzryclfequ5isakqorz6oedvq2vl6cjektcpmmt7fm",
						"height": int64(4),
					},
					{
						"cid":    "bafybeibvbwkak42qyfgwg6rnlfelhovriqucsuq2kz77d6z7h7m46k4sdi",
						"height": int64(4),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
