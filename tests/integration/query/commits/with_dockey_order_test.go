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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeibvzg7f2p772ev3srlzt4w5jjwlo3nw4chtd6ewuvbrnlidzqtmr4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiahsvsfxvytbmyek7mjzh666y2qz2jlfse4fdgwzx4lnunuukurcm",
						"height": int64(2),
					},
					{
						"cid":    "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
						"height": int64(1),
					},
					{
						"cid":    "bafybeifukwb3t73k7pph3ctp5khosoycp53ywjl6btravzk6decggkjtl4",
						"height": int64(1),
					},
					{
						"cid":    "bafybeig3wrpwi6q7vjchizcwnenslasyxop6wey7jahbiszlubdglfq2fq",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}

func TestQueryCommitsWithDockeyAndOrderHeightAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order height asc",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
						"height": int64(1),
					},
					{
						"cid":    "bafybeifukwb3t73k7pph3ctp5khosoycp53ywjl6btravzk6decggkjtl4",
						"height": int64(1),
					},
					{
						"cid":    "bafybeig3wrpwi6q7vjchizcwnenslasyxop6wey7jahbiszlubdglfq2fq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeibvzg7f2p772ev3srlzt4w5jjwlo3nw4chtd6ewuvbrnlidzqtmr4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiahsvsfxvytbmyek7mjzh666y2qz2jlfse4fdgwzx4lnunuukurcm",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}

func TestQueryCommitsWithDockeyAndOrderCidDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order cid desc",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeig3wrpwi6q7vjchizcwnenslasyxop6wey7jahbiszlubdglfq2fq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeifukwb3t73k7pph3ctp5khosoycp53ywjl6btravzk6decggkjtl4",
						"height": int64(1),
					},
					{
						"cid":    "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
						"height": int64(1),
					},
					{
						"cid":    "bafybeibvzg7f2p772ev3srlzt4w5jjwlo3nw4chtd6ewuvbrnlidzqtmr4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiahsvsfxvytbmyek7mjzh666y2qz2jlfse4fdgwzx4lnunuukurcm",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}

func TestQueryCommitsWithDockeyAndOrderCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order cid asc",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeiahsvsfxvytbmyek7mjzh666y2qz2jlfse4fdgwzx4lnunuukurcm",
						"height": int64(2),
					},
					{
						"cid":    "bafybeibvzg7f2p772ev3srlzt4w5jjwlo3nw4chtd6ewuvbrnlidzqtmr4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
						"height": int64(1),
					},
					{
						"cid":    "bafybeifukwb3t73k7pph3ctp5khosoycp53ywjl6btravzk6decggkjtl4",
						"height": int64(1),
					},
					{
						"cid":    "bafybeig3wrpwi6q7vjchizcwnenslasyxop6wey7jahbiszlubdglfq2fq",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}

func TestQueryCommitsWithDockeyAndOrderAndMultiUpdatesCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple updates with order cid asc",
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
						 commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							 cid
							 height
						 }
					 }`,
				Results: []map[string]any{
					{
						"cid":    "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
						"height": int64(1),
					},
					{
						"cid":    "bafybeifukwb3t73k7pph3ctp5khosoycp53ywjl6btravzk6decggkjtl4",
						"height": int64(1),
					},
					{
						"cid":    "bafybeig3wrpwi6q7vjchizcwnenslasyxop6wey7jahbiszlubdglfq2fq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeibvzg7f2p772ev3srlzt4w5jjwlo3nw4chtd6ewuvbrnlidzqtmr4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiahsvsfxvytbmyek7mjzh666y2qz2jlfse4fdgwzx4lnunuukurcm",
						"height": int64(2),
					},
					{
						"cid":    "bafybeihccn3utqsaxzsh6i7dlnd45rutcg7fbsogfw4vvigii7laedslqe",
						"height": int64(3),
					},
					{
						"cid":    "bafybeiegusf5ypa7htxwa6u4fvne3lqq2jafe4fxllh4lo6iw4xdsn4yyq",
						"height": int64(3),
					},
					{
						"cid":    "bafybeigicex7hqzhzltm3adsx34rnzhp7lgubtrusxukk54whosmtfun7y",
						"height": int64(4),
					},
					{
						"cid":    "bafybeihv6d4fo7q5pziriv4rz3loq6unr3fegdonjcuyw5stano5r7dm4i",
						"height": int64(4),
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}
