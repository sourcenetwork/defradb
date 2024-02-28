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
						"cid":    "bafybeicwg56ddi7smy3j2kkv5y4yghvdrj3twqqafzdwtinbkw5mlpxwz4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidxnkwhuzmkdw5wuippru3tp74vcmz5jvcziambpjadxeathdh26a",
						"height": int64(2),
					},
					{
						"cid":    "bafybeicvpgfinf2m2jufbbcy5mhv6jca6in5k4fzx5op7xvvcmbp7sceaa",
						"height": int64(1),
					},
					{
						"cid":    "bafybeib2espk2hq366wjnmazg45uvoswqbvf4plx7fgzayagxdn737onci",
						"height": int64(1),
					},
					{
						"cid":    "bafybeigvpf62j7j2wbpid5iavzxielbhbsbbirmgzqkw3wpptdvysuztwi",
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
						"cid":    "bafybeicvpgfinf2m2jufbbcy5mhv6jca6in5k4fzx5op7xvvcmbp7sceaa",
						"height": int64(1),
					},
					{
						"cid":    "bafybeib2espk2hq366wjnmazg45uvoswqbvf4plx7fgzayagxdn737onci",
						"height": int64(1),
					},
					{
						"cid":    "bafybeigvpf62j7j2wbpid5iavzxielbhbsbbirmgzqkw3wpptdvysuztwi",
						"height": int64(1),
					},
					{
						"cid":    "bafybeicwg56ddi7smy3j2kkv5y4yghvdrj3twqqafzdwtinbkw5mlpxwz4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidxnkwhuzmkdw5wuippru3tp74vcmz5jvcziambpjadxeathdh26a",
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
						"cid":    "bafybeigvpf62j7j2wbpid5iavzxielbhbsbbirmgzqkw3wpptdvysuztwi",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidxnkwhuzmkdw5wuippru3tp74vcmz5jvcziambpjadxeathdh26a",
						"height": int64(2),
					},
					{
						"cid":    "bafybeicwg56ddi7smy3j2kkv5y4yghvdrj3twqqafzdwtinbkw5mlpxwz4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeicvpgfinf2m2jufbbcy5mhv6jca6in5k4fzx5op7xvvcmbp7sceaa",
						"height": int64(1),
					},
					{
						"cid":    "bafybeib2espk2hq366wjnmazg45uvoswqbvf4plx7fgzayagxdn737onci",
						"height": int64(1),
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
						"cid":    "bafybeib2espk2hq366wjnmazg45uvoswqbvf4plx7fgzayagxdn737onci",
						"height": int64(1),
					},
					{
						"cid":    "bafybeicvpgfinf2m2jufbbcy5mhv6jca6in5k4fzx5op7xvvcmbp7sceaa",
						"height": int64(1),
					},
					{
						"cid":    "bafybeicwg56ddi7smy3j2kkv5y4yghvdrj3twqqafzdwtinbkw5mlpxwz4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidxnkwhuzmkdw5wuippru3tp74vcmz5jvcziambpjadxeathdh26a",
						"height": int64(2),
					},
					{
						"cid":    "bafybeigvpf62j7j2wbpid5iavzxielbhbsbbirmgzqkw3wpptdvysuztwi",
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
						"cid":    "bafybeicvpgfinf2m2jufbbcy5mhv6jca6in5k4fzx5op7xvvcmbp7sceaa",
						"height": int64(1),
					},
					{
						"cid":    "bafybeib2espk2hq366wjnmazg45uvoswqbvf4plx7fgzayagxdn737onci",
						"height": int64(1),
					},
					{
						"cid":    "bafybeigvpf62j7j2wbpid5iavzxielbhbsbbirmgzqkw3wpptdvysuztwi",
						"height": int64(1),
					},
					{
						"cid":    "bafybeicwg56ddi7smy3j2kkv5y4yghvdrj3twqqafzdwtinbkw5mlpxwz4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidxnkwhuzmkdw5wuippru3tp74vcmz5jvcziambpjadxeathdh26a",
						"height": int64(2),
					},
					{
						"cid":    "bafybeic45wkhxtpn3vgd2dmmohq76vw56qz3cpu3oorha3hf2w6qu7bpoa",
						"height": int64(3),
					},
					{
						"cid":    "bafybeid4y4vqmvec2mvm3su77rrmj6tzsx5zdlt6ias4hzqxbevmosydc4",
						"height": int64(3),
					},
					{
						"cid":    "bafybeiatfviresatclvedt6zhk4ys7p6cdts5udqsl33nu5d2hxtw4l6la",
						"height": int64(4),
					},
					{
						"cid":    "bafybeiaydxxf7bmeh5ou47z6exa73heg6vjjzznbvrxqbemmu55sdhvuom",
						"height": int64(4),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
