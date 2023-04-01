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
						"cid":    "bafybeiepww5b67jrrliuiy27erfjuivwnjca5ptdpbxrrjrqkh3b2hckyy",
						"height": int64(2),
					},
					{
						"cid":    "bafybeigk3hiskvcflfgk5cuivr7dnwhwsjlh3vyfkqxyvkpnz72kf5zn7i",
						"height": int64(2),
					},
					{
						"cid":    "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
						"height": int64(1),
					},
					{
						"cid":    "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"height": int64(1),
					},
					{
						"cid":    "bafybeicjgxd46coc73ybx7cvdxk5c6pdbp5nbick4drrlbevgg7a7mhkzy",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
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
						"cid":    "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
						"height": int64(1),
					},
					{
						"cid":    "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"height": int64(1),
					},
					{
						"cid":    "bafybeicjgxd46coc73ybx7cvdxk5c6pdbp5nbick4drrlbevgg7a7mhkzy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiepww5b67jrrliuiy27erfjuivwnjca5ptdpbxrrjrqkh3b2hckyy",
						"height": int64(2),
					},
					{
						"cid":    "bafybeigk3hiskvcflfgk5cuivr7dnwhwsjlh3vyfkqxyvkpnz72kf5zn7i",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
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
						"cid":    "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
						"height": int64(1),
					},
					{
						"cid":    "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"height": int64(1),
					},
					{
						"cid":    "bafybeigk3hiskvcflfgk5cuivr7dnwhwsjlh3vyfkqxyvkpnz72kf5zn7i",
						"height": int64(2),
					},
					{
						"cid":    "bafybeiepww5b67jrrliuiy27erfjuivwnjca5ptdpbxrrjrqkh3b2hckyy",
						"height": int64(2),
					},
					{
						"cid":    "bafybeicjgxd46coc73ybx7cvdxk5c6pdbp5nbick4drrlbevgg7a7mhkzy",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
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
						"cid":    "bafybeicjgxd46coc73ybx7cvdxk5c6pdbp5nbick4drrlbevgg7a7mhkzy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiepww5b67jrrliuiy27erfjuivwnjca5ptdpbxrrjrqkh3b2hckyy",
						"height": int64(2),
					},
					{
						"cid":    "bafybeigk3hiskvcflfgk5cuivr7dnwhwsjlh3vyfkqxyvkpnz72kf5zn7i",
						"height": int64(2),
					},
					{
						"cid":    "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"height": int64(1),
					},
					{
						"cid":    "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
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
						"cid":    "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
						"height": int64(1),
					},
					{
						"cid":    "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"height": int64(1),
					},
					{
						"cid":    "bafybeicjgxd46coc73ybx7cvdxk5c6pdbp5nbick4drrlbevgg7a7mhkzy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiepww5b67jrrliuiy27erfjuivwnjca5ptdpbxrrjrqkh3b2hckyy",
						"height": int64(2),
					},
					{
						"cid":    "bafybeigk3hiskvcflfgk5cuivr7dnwhwsjlh3vyfkqxyvkpnz72kf5zn7i",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidag5ogxiqsctdxktobw6gbq52pbfkmicp6np5ccuy6pjyvgobdxq",
						"height": int64(3),
					},
					{
						"cid":    "bafybeihjmxcpf7ab7uzktqs7d7mt5hytrkchzsqz4j2ab3gwhqlpulbfsu",
						"height": int64(3),
					},
					{
						"cid":    "bafybeiadnlnuu3gkfaf5fgib3iqfsntitb4pjrc5hffvmhpdjqgsxutc5q",
						"height": int64(4),
					},
					{
						"cid":    "bafybeicnmkwcoopkwy2cah3jyxmxjvsp65sl2e6d5mqb5ygfdyjvfbydhi",
						"height": int64(4),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}
