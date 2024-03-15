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
						"cid":    "bafybeia7qkfbfm4jijlkqs6uxziie2v57nin5gaa3afnpkruw352mmrt4q",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeih4plbb3rinhqvn663ssfwhnujdbnbjistymzowsry5nvmxchmqny",
						"height": int64(1),
					},
					{
						"cid":    "bafybeid4fh7ggr2wgema6b5hrqroimcso3vxyous3oyck5c66vm72br7z4",
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
						"cid":    "bafybeiewgawahat7sxdoafvu77uvsaaj2ttatqllj2qvnqhornzxl2gteq",
						"height": int64(3),
					},
					{
						// Composite head -1
						"cid":    "bafybeia7qkfbfm4jijlkqs6uxziie2v57nin5gaa3afnpkruw352mmrt4q",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeih4plbb3rinhqvn663ssfwhnujdbnbjistymzowsry5nvmxchmqny",
						"height": int64(1),
					},
					{
						// "Age" field head
						"cid":    "bafybeid44afmsi6hh6yasgcjncnlvdpqsu2durizsxhdmhsbrqekypf6aa",
						"height": int64(3),
					},
					{
						// "Age" field head -1
						"cid":    "bafybeid4fh7ggr2wgema6b5hrqroimcso3vxyous3oyck5c66vm72br7z4",
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
						"cid": "bafybeieepmzk3s5dzxztfq5zi5e5g3mnb6yfumx3euknnbur4x3a5neidq",
					},
					{
						"cid": "bafybeifcxdrzqfj54w5mls7mf6nhxtjnweoevves7rwzsda6gmvzqc4t7y",
					},
					{
						"cid": "bafybeihavavtkfgaevtnzbabdwmgpamgbpkonw4ardalsamcitspqaxhs4",
					},
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
