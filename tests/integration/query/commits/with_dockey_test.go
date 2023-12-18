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

func TestQueryCommitsWithUnknownDockey(t *testing.T) {
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
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDockey(t *testing.T) {
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
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

func TestQueryCommitsWithDockeyAndLinks(t *testing.T) {
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid":   "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeib5dcp5ccsm3r3sey6td23r7qgd2xjqnx4gfxss6hd77zhjl44gbq",
						"links": []map[string]any{
							{
								"cid":  "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
								"name": "age",
							},
							{
								"cid":  "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
								"name": "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDockeyAndUpdate(t *testing.T) {
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
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
						"cid":    "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
						"height": int64(1),
					},
					{
						"cid":    "bafybeievtwczfax2ncnnoc72bnt4drl5vtz3qaqgl3odwfxzwjkfsa2zey",
						"height": int64(2),
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

// This test is for documentation reasons only. This is not
// desired behaviour (first results includes link._head, second
// includes link._Name).
func TestQueryCommitsWithDockeyAndUpdateAndLinks(t *testing.T) {
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
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeifkbeua3tz2oao2dclsphbrszpmt7t5m66y76bxnjssdeuztrsjtm",
						"links": []map[string]any{
							{
								"cid":  "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
								"name": "_head",
							},
						},
					},
					{
						"cid":   "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeievtwczfax2ncnnoc72bnt4drl5vtz3qaqgl3odwfxzwjkfsa2zey",
						"links": []map[string]any{
							{
								"cid":  "bafybeib5dcp5ccsm3r3sey6td23r7qgd2xjqnx4gfxss6hd77zhjl44gbq",
								"name": "_head",
							},
							{
								"cid":  "bafybeifkbeua3tz2oao2dclsphbrszpmt7t5m66y76bxnjssdeuztrsjtm",
								"name": "age",
							},
						},
					},
					{
						"cid": "bafybeib5dcp5ccsm3r3sey6td23r7qgd2xjqnx4gfxss6hd77zhjl44gbq",
						"links": []map[string]any{
							{
								"cid":  "bafybeiaayywpuodsyqnp6vbnftzpzoke4iwofeyuta3wyfv6xpjenf7le4",
								"name": "age",
							},
							{
								"cid":  "bafybeiam53sbmu6zqvullm53sq2hlrdc4icifbcznmltg5olijyrdnjhtm",
								"name": "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
