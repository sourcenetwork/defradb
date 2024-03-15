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

func TestQueryCommitsWithUnknownDocID(t *testing.T) {
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

func TestQueryCommitsWithDocID(t *testing.T) {
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
						"cid": "bafybeieoeoset5itv7alud2yzjmq6dqizymdwdmlvyxam2uxe4lfexooaq",
					},
					{
						"cid": "bafybeih4plbb3rinhqvn663ssfwhnujdbnbjistymzowsry5nvmxchmqny",
					},
					{
						"cid": "bafybeic2zvs2beirqmgd45myszkqwj32w3oyduolugkxv4gxxph4c4mzva",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndLinks(t *testing.T) {
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
						"cid":   "bafybeieoeoset5itv7alud2yzjmq6dqizymdwdmlvyxam2uxe4lfexooaq",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeih4plbb3rinhqvn663ssfwhnujdbnbjistymzowsry5nvmxchmqny",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeic2zvs2beirqmgd45myszkqwj32w3oyduolugkxv4gxxph4c4mzva",
						"links": []map[string]any{
							{
								"cid":  "bafybeieoeoset5itv7alud2yzjmq6dqizymdwdmlvyxam2uxe4lfexooaq",
								"name": "age",
							},
							{
								"cid":  "bafybeih4plbb3rinhqvn663ssfwhnujdbnbjistymzowsry5nvmxchmqny",
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

func TestQueryCommitsWithDocIDAndUpdate(t *testing.T) {
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
						"cid":    "bafybeia7qkfbfm4jijlkqs6uxziie2v57nin5gaa3afnpkruw352mmrt4q",
						"height": int64(2),
					},
					{
						"cid":    "bafybeieoeoset5itv7alud2yzjmq6dqizymdwdmlvyxam2uxe4lfexooaq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeih4plbb3rinhqvn663ssfwhnujdbnbjistymzowsry5nvmxchmqny",
						"height": int64(1),
					},
					{
						"cid":    "bafybeid4fh7ggr2wgema6b5hrqroimcso3vxyous3oyck5c66vm72br7z4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeic2zvs2beirqmgd45myszkqwj32w3oyduolugkxv4gxxph4c4mzva",
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
func TestQueryCommitsWithDocIDAndUpdateAndLinks(t *testing.T) {
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
						"cid": "bafybeia7qkfbfm4jijlkqs6uxziie2v57nin5gaa3afnpkruw352mmrt4q",
						"links": []map[string]any{
							{
								"cid":  "bafybeieoeoset5itv7alud2yzjmq6dqizymdwdmlvyxam2uxe4lfexooaq",
								"name": "_head",
							},
						},
					},
					{
						"cid":   "bafybeieoeoset5itv7alud2yzjmq6dqizymdwdmlvyxam2uxe4lfexooaq",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeih4plbb3rinhqvn663ssfwhnujdbnbjistymzowsry5nvmxchmqny",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeid4fh7ggr2wgema6b5hrqroimcso3vxyous3oyck5c66vm72br7z4",
						"links": []map[string]any{
							{
								"cid":  "bafybeic2zvs2beirqmgd45myszkqwj32w3oyduolugkxv4gxxph4c4mzva",
								"name": "_head",
							},
							{
								"cid":  "bafybeia7qkfbfm4jijlkqs6uxziie2v57nin5gaa3afnpkruw352mmrt4q",
								"name": "age",
							},
						},
					},
					{
						"cid": "bafybeic2zvs2beirqmgd45myszkqwj32w3oyduolugkxv4gxxph4c4mzva",
						"links": []map[string]any{
							{
								"cid":  "bafybeieoeoset5itv7alud2yzjmq6dqizymdwdmlvyxam2uxe4lfexooaq",
								"name": "age",
							},
							{
								"cid":  "bafybeih4plbb3rinhqvn663ssfwhnujdbnbjistymzowsry5nvmxchmqny",
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
