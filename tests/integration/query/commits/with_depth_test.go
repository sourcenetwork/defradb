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
						"Name":	"John",
						"Age":	21
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
						"cid": "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
					},
					{
						"cid": "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
					},
					{
						"cid": "bafybeicjgxd46coc73ybx7cvdxk5c6pdbp5nbick4drrlbevgg7a7mhkzy",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithDepth1WithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with depth 1, and doc updates",
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
						commits(depth: 1) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						// "Age" field head
						"cid":    "bafybeiepww5b67jrrliuiy27erfjuivwnjca5ptdpbxrrjrqkh3b2hckyy",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"height": int64(1),
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

func TestQueryCommitsWithDepth2WithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with depth 2, and doc updates",
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
						"cid":    "bafybeidag5ogxiqsctdxktobw6gbq52pbfkmicp6np5ccuy6pjyvgobdxq",
						"height": int64(3),
					},
					{
						// Composite head -1
						"cid":    "bafybeiepww5b67jrrliuiy27erfjuivwnjca5ptdpbxrrjrqkh3b2hckyy",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"height": int64(1),
					},
					{
						// "Age" field head
						"cid":    "bafybeihjmxcpf7ab7uzktqs7d7mt5hytrkchzsqz4j2ab3gwhqlpulbfsu",
						"height": int64(3),
					},
					{
						// "Age" field head -1
						"cid":    "bafybeigk3hiskvcflfgk5cuivr7dnwhwsjlh3vyfkqxyvkpnz72kf5zn7i",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithDepth1AndMultipleDocs(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with depth 1",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"Name":	"Fred",
						"Age":	25
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
						"cid": "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
					},
					{
						"cid": "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
					},
					{
						"cid": "bafybeicjgxd46coc73ybx7cvdxk5c6pdbp5nbick4drrlbevgg7a7mhkzy",
					},
					{
						"cid": "bafybeifo3ujh2imjbx7d4g43eocureejdcclb2zpkzelwtt2omaxzxkxlm",
					},
					{
						"cid": "bafybeih2tjbbltowuprpc7q7tx7jw7zwlso5s3yck3o5gmi3dsr54ooyey",
					},
					{
						"cid": "bafybeih2fsuzskmm6vlk2t274hw2mzitye6bbc4sozppokhhy4piphn6qq",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}
