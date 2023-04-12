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

func TestQueryCommitsWithDockeyAndOrderHeightDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order height desc",
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
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {height: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiculvxijb7chmmmb2cad3k4e2zfug5tpehp4gqwayarh4yzlnft4m",
						"height": int64(2),
					},
					{
						"cid":    "bafybeifh5fpsnusu6tedflmcglsm4rxsc2k3h35iqsbrcevfish6xrxftq",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiezqs7t2piarfu3g7s3bmjzok7zm4tpc5tgbp4sovdfvggrjfpjxa",
						"height": int64(1),
					},
					{
						"cid":    "bafybeid32koipavfiznapnb7cbhbwqfi7vmbob6xj6tzj2w66utir4qv5y",
						"height": int64(1),
					},
					{
						"cid":    "bafybeickezedvwqujfvdih25kcwpkxtekdy3ykaofxejtxmhqybw2pfrnq",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestQueryCommitsWithDockeyAndOrderHeightAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order height asc",
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
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {height: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiezqs7t2piarfu3g7s3bmjzok7zm4tpc5tgbp4sovdfvggrjfpjxa",
						"height": int64(1),
					},
					{
						"cid":    "bafybeid32koipavfiznapnb7cbhbwqfi7vmbob6xj6tzj2w66utir4qv5y",
						"height": int64(1),
					},
					{
						"cid":    "bafybeickezedvwqujfvdih25kcwpkxtekdy3ykaofxejtxmhqybw2pfrnq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiculvxijb7chmmmb2cad3k4e2zfug5tpehp4gqwayarh4yzlnft4m",
						"height": int64(2),
					},
					{
						"cid":    "bafybeifh5fpsnusu6tedflmcglsm4rxsc2k3h35iqsbrcevfish6xrxftq",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestQueryCommitsWithDockeyAndOrderCidDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order cid desc",
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
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {cid: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeifh5fpsnusu6tedflmcglsm4rxsc2k3h35iqsbrcevfish6xrxftq",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiezqs7t2piarfu3g7s3bmjzok7zm4tpc5tgbp4sovdfvggrjfpjxa",
						"height": int64(1),
					},
					{
						"cid":    "bafybeid32koipavfiznapnb7cbhbwqfi7vmbob6xj6tzj2w66utir4qv5y",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiculvxijb7chmmmb2cad3k4e2zfug5tpehp4gqwayarh4yzlnft4m",
						"height": int64(2),
					},
					{
						"cid":    "bafybeickezedvwqujfvdih25kcwpkxtekdy3ykaofxejtxmhqybw2pfrnq",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestQueryCommitsWithDockeyAndOrderCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order cid asc",
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
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {cid: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeickezedvwqujfvdih25kcwpkxtekdy3ykaofxejtxmhqybw2pfrnq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiezqs7t2piarfu3g7s3bmjzok7zm4tpc5tgbp4sovdfvggrjfpjxa",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiculvxijb7chmmmb2cad3k4e2zfug5tpehp4gqwayarh4yzlnft4m",
						"height": int64(2),
					},
					{
						"cid":    "bafybeid32koipavfiznapnb7cbhbwqfi7vmbob6xj6tzj2w66utir4qv5y",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiezqs7t2piarfu3g7s3bmjzok7zm4tpc5tgbp4sovdfvggrjfpjxa",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestQueryCommitsWithDockeyAndOrderAndMultiUpdatesCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple updates with order cid asc",
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
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Age":	23
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"Age":	24
				}`,
			},
			testUtils.Request{
				Request: `query {
						 commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f", order: {height: ASC}) {
							 cid
							 height
						 }
					 }`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiezqs7t2piarfu3g7s3bmjzok7zm4tpc5tgbp4sovdfvggrjfpjxa",
						"height": int64(1),
					},
					{
						"cid":    "bafybeid32koipavfiznapnb7cbhbwqfi7vmbob6xj6tzj2w66utir4qv5y",
						"height": int64(1),
					},
					{
						"cid":    "bafybeickezedvwqujfvdih25kcwpkxtekdy3ykaofxejtxmhqybw2pfrnq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiculvxijb7chmmmb2cad3k4e2zfug5tpehp4gqwayarh4yzlnft4m",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiezqs7t2piarfu3g7s3bmjzok7zm4tpc5tgbp4sovdfvggrjfpjxa",
						"height": int64(2),
					},
					{
						"cid":    "bafybeibycdufs2ltfbhqqo6baq2bndj3b56bjb4fzijm3st7dbflbj5vhy",
						"height": int64(3),
					},
					{
						"cid":    "bafybeihbo6inadk6xuaashol2yahu474o7ns7s7ploq55a2qcngha6czde",
						"height": int64(3),
					},
					{
						"cid":    "bafybeidqprfw43ydrkla7rkzoeyg5773vumgylxmg7rvabo4eb62mgdsmu",
						"height": int64(4),
					},
					{
						"cid":    "bafybeicjqj6eax5otkqcydkrr5nvxwm25pxntnpwqk3mv2bvk4a4ntxsau",
						"height": int64(4),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
