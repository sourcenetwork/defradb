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
		Description: "Simple all commits query with unknown dockey",
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
						commits(dockey: "unknown dockey") {
							cid
						}
					}`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}

func TestQueryCommitsWithDockey(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
					},
					{
						"cid": "bafybeifukwb3t73k7pph3ctp5khosoycp53ywjl6btravzk6decggkjtl4",
					},
					{
						"cid": "bafybeig3wrpwi6q7vjchizcwnenslasyxop6wey7jahbiszlubdglfq2fq",
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}

func TestQueryCommitsWithDockeyAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, with links",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid":   "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeifukwb3t73k7pph3ctp5khosoycp53ywjl6btravzk6decggkjtl4",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeig3wrpwi6q7vjchizcwnenslasyxop6wey7jahbiszlubdglfq2fq",
						"links": []map[string]any{
							{
								"cid":  "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
								"name": "age",
							},
							{
								"cid":  "bafybeifukwb3t73k7pph3ctp5khosoycp53ywjl6btravzk6decggkjtl4",
								"name": "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}

func TestQueryCommitsWithDockeyAndUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple results",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
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
						"cid":    "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
						"height": int64(1),
					},
					{
						"cid":    "bafybeifukwb3t73k7pph3ctp5khosoycp53ywjl6btravzk6decggkjtl4",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiahsvsfxvytbmyek7mjzh666y2qz2jlfse4fdgwzx4lnunuukurcm",
						"height": int64(2),
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

// This test is for documentation reasons only. This is not
// desired behaviour (first results includes link._head, second
// includes link._Name).
func TestQueryCommitsWithDockeyAndUpdateAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple results and links",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeibvzg7f2p772ev3srlzt4w5jjwlo3nw4chtd6ewuvbrnlidzqtmr4",
						"links": []map[string]any{
							{
								"cid":  "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
								"name": "_head",
							},
						},
					},
					{
						"cid":   "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeifukwb3t73k7pph3ctp5khosoycp53ywjl6btravzk6decggkjtl4",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeiahsvsfxvytbmyek7mjzh666y2qz2jlfse4fdgwzx4lnunuukurcm",
						"links": []map[string]any{
							{
								"cid":  "bafybeig3wrpwi6q7vjchizcwnenslasyxop6wey7jahbiszlubdglfq2fq",
								"name": "_head",
							},
							{
								"cid":  "bafybeibvzg7f2p772ev3srlzt4w5jjwlo3nw4chtd6ewuvbrnlidzqtmr4",
								"name": "age",
							},
						},
					},
					{
						"cid": "bafybeig3wrpwi6q7vjchizcwnenslasyxop6wey7jahbiszlubdglfq2fq",
						"links": []map[string]any{
							{
								"cid":  "bafybeic5oodfpnixl6uf4bi63m3eouuhj3gafudlsd4tqryhx2wy7rczoe",
								"name": "age",
							},
							{
								"cid":  "bafybeifukwb3t73k7pph3ctp5khosoycp53ywjl6btravzk6decggkjtl4",
								"name": "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTEMP(t, test)
}
