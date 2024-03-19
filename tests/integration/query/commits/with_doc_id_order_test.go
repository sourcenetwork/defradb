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
						"cid":    "bafybeia7qkfbfm4jijlkqs6uxziie2v57nin5gaa3afnpkruw352mmrt4q",
						"height": int64(2),
					},
					{
						"cid":    "bafybeid4fh7ggr2wgema6b5hrqroimcso3vxyous3oyck5c66vm72br7z4",
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
						"cid":    "bafybeic2zvs2beirqmgd45myszkqwj32w3oyduolugkxv4gxxph4c4mzva",
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
						"cid":    "bafybeieoeoset5itv7alud2yzjmq6dqizymdwdmlvyxam2uxe4lfexooaq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeih4plbb3rinhqvn663ssfwhnujdbnbjistymzowsry5nvmxchmqny",
						"height": int64(1),
					},
					{
						"cid":    "bafybeic2zvs2beirqmgd45myszkqwj32w3oyduolugkxv4gxxph4c4mzva",
						"height": int64(1),
					},
					{
						"cid":    "bafybeia7qkfbfm4jijlkqs6uxziie2v57nin5gaa3afnpkruw352mmrt4q",
						"height": int64(2),
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
						"cid":    "bafybeih4plbb3rinhqvn663ssfwhnujdbnbjistymzowsry5nvmxchmqny",
						"height": int64(1),
					},
					{
						"cid":    "bafybeieoeoset5itv7alud2yzjmq6dqizymdwdmlvyxam2uxe4lfexooaq",
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
					{
						"cid":    "bafybeia7qkfbfm4jijlkqs6uxziie2v57nin5gaa3afnpkruw352mmrt4q",
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
						"cid":    "bafybeia7qkfbfm4jijlkqs6uxziie2v57nin5gaa3afnpkruw352mmrt4q",
						"height": int64(2),
					},
					{
						"cid":    "bafybeic2zvs2beirqmgd45myszkqwj32w3oyduolugkxv4gxxph4c4mzva",
						"height": int64(1),
					},
					{
						"cid":    "bafybeid4fh7ggr2wgema6b5hrqroimcso3vxyous3oyck5c66vm72br7z4",
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
						"cid":    "bafybeieoeoset5itv7alud2yzjmq6dqizymdwdmlvyxam2uxe4lfexooaq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeih4plbb3rinhqvn663ssfwhnujdbnbjistymzowsry5nvmxchmqny",
						"height": int64(1),
					},
					{
						"cid":    "bafybeic2zvs2beirqmgd45myszkqwj32w3oyduolugkxv4gxxph4c4mzva",
						"height": int64(1),
					},
					{
						"cid":    "bafybeia7qkfbfm4jijlkqs6uxziie2v57nin5gaa3afnpkruw352mmrt4q",
						"height": int64(2),
					},
					{
						"cid":    "bafybeid4fh7ggr2wgema6b5hrqroimcso3vxyous3oyck5c66vm72br7z4",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiewgawahat7sxdoafvu77uvsaaj2ttatqllj2qvnqhornzxl2gteq",
						"height": int64(3),
					},
					{
						"cid":    "bafybeid44afmsi6hh6yasgcjncnlvdpqsu2durizsxhdmhsbrqekypf6aa",
						"height": int64(3),
					},
					{
						"cid":    "bafybeifq2bd3nkqa6q5tjb5lrmeoskimtchhodvcxdqeilck2x4k3z7ijq",
						"height": int64(4),
					},
					{
						"cid":    "bafybeiakoro6m2bvfmtmczykyffvixx6ci7tbgczebhdwttx5ymh5c7wyy",
						"height": int64(4),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
