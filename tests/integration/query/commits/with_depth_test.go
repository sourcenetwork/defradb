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
						"cid": "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
					},
					{
						"cid": "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
					},
					{
						"cid": "bafybeib5dcp5ccsm3r3sey6td23r7qgd2xjqnx4gfxss6hd77zhjl44gbq",
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
						"cid":    "bafybeifkbeua3tz2oao2dclsphbrszpmt7t5m66y76bxnjssdeuztrsjtm",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
						"height": int64(1),
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
						"cid":    "bafybeicdu5nwxxnvm7ssnpjbzkad435csokotelgqpkebbppm4etxqmmoq",
						"height": int64(3),
					},
					{
						// Composite head -1
						"cid":    "bafybeifkbeua3tz2oao2dclsphbrszpmt7t5m66y76bxnjssdeuztrsjtm",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
						"height": int64(1),
					},
					{
						// "Age" field head
						"cid":    "bafybeici4m36wbsbw2u4yg64udxjxwul3k5bssin6zboqxrabbqamoc53u",
						"height": int64(3),
					},
					{
						// "Age" field head -1
						"cid":    "bafybeievtwczfax2ncnnoc72bnt4drl5vtz3qaqgl3odwfxzwjkfsa2zey",
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
						"cid": "bafybeifxfb7gy5ugfwku6m6k2l46fqov47qatye577s6rwzn4dk5ngfv3q",
					},
					{
						"cid": "bafybeig3g2zmjgn7ggo2y3vvqvkntbclkevxva3ata7cymal3jor4bn45a",
					},
					{
						"cid": "bafybeidtnj57krjiwedp7gcyyloacyvvqdnb3pmvitveymyy5gvajm26fm",
					},
					{
						"cid": "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
					},
					{
						"cid": "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
					},
					{
						"cid": "bafybeib5dcp5ccsm3r3sey6td23r7qgd2xjqnx4gfxss6hd77zhjl44gbq",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
