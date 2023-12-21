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
						"cid":    "bafybeifkbeua3tz2oao2dclsphbrszpmt7t5m66y76bxnjssdeuztrsjtm",
						"height": int64(2),
					},
					{
						"cid":    "bafybeievtwczfax2ncnnoc72bnt4drl5vtz3qaqgl3odwfxzwjkfsa2zey",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
						"height": int64(1),
					},
					{
						"cid":    "bafybeib5dcp5ccsm3r3sey6td23r7qgd2xjqnx4gfxss6hd77zhjl44gbq",
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
						"cid":    "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
						"height": int64(1),
					},
					{
						"cid":    "bafybeib5dcp5ccsm3r3sey6td23r7qgd2xjqnx4gfxss6hd77zhjl44gbq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeifkbeua3tz2oao2dclsphbrszpmt7t5m66y76bxnjssdeuztrsjtm",
						"height": int64(2),
					},
					{
						"cid":    "bafybeievtwczfax2ncnnoc72bnt4drl5vtz3qaqgl3odwfxzwjkfsa2zey",
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
						"cid":    "bafybeifkbeua3tz2oao2dclsphbrszpmt7t5m66y76bxnjssdeuztrsjtm",
						"height": int64(2),
					},
					{
						"cid":    "bafybeievtwczfax2ncnnoc72bnt4drl5vtz3qaqgl3odwfxzwjkfsa2zey",
						"height": int64(2),
					},
					{
						"cid":    "bafybeib5dcp5ccsm3r3sey6td23r7qgd2xjqnx4gfxss6hd77zhjl44gbq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
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
						"cid":    "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
						"height": int64(1),
					},
					{
						"cid":    "bafybeib5dcp5ccsm3r3sey6td23r7qgd2xjqnx4gfxss6hd77zhjl44gbq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeievtwczfax2ncnnoc72bnt4drl5vtz3qaqgl3odwfxzwjkfsa2zey",
						"height": int64(2),
					},
					{
						"cid":    "bafybeifkbeua3tz2oao2dclsphbrszpmt7t5m66y76bxnjssdeuztrsjtm",
						"height": int64(2),
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
						"cid":    "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
						"height": int64(1),
					},
					{
						"cid":    "bafybeib5dcp5ccsm3r3sey6td23r7qgd2xjqnx4gfxss6hd77zhjl44gbq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeifkbeua3tz2oao2dclsphbrszpmt7t5m66y76bxnjssdeuztrsjtm",
						"height": int64(2),
					},
					{
						"cid":    "bafybeievtwczfax2ncnnoc72bnt4drl5vtz3qaqgl3odwfxzwjkfsa2zey",
						"height": int64(2),
					},
					{
						"cid":    "bafybeicdu5nwxxnvm7ssnpjbzkad435csokotelgqpkebbppm4etxqmmoq",
						"height": int64(3),
					},
					{
						"cid":    "bafybeici4m36wbsbw2u4yg64udxjxwul3k5bssin6zboqxrabbqamoc53u",
						"height": int64(3),
					},
					{
						"cid":    "bafybeiba44cladlw2mok3yk5ayhwqmpztq2pyxav5tjaprxzcmjhcjdt74",
						"height": int64(4),
					},
					{
						"cid":    "bafybeibcndl6uyim7u573mxcrdi6iui6y5c35cqejpic4ife67wjoxzknq",
						"height": int64(4),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
