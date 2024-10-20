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
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreif6dqbkr7t37jcjfxxrjnxt7cspxzvs7qwlbtjca57cc663he4s7e",
						},
						{
							"cid": "bafyreigtnj6ntulcilkmin4pgukjwv3nwglqpiiyddz3dyfexdbltze7sy",
						},
						{
							"cid": "bafyreia2vlbfkcbyogdjzmbqcjneabwwwtw7ti2xbd7yor5mbu2sk4pcoy",
						},
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
				Results: map[string]any{
					"commits": []map[string]any{
						{
							// "Age" field head
							"cid":    "bafyreih5h6i6ohfsgrcjtg76iarebqcurpaft73gpobl2z2cfsvihsgdqu",
							"height": int64(2),
						},
						{
							// "Name" field head (unchanged from create)
							"cid":    "bafyreigtnj6ntulcilkmin4pgukjwv3nwglqpiiyddz3dyfexdbltze7sy",
							"height": int64(1),
						},
						{
							"cid":    "bafyreiale6qsjc7qewod3c6h2odwamfwcf7vt4zlqtw7ldcm57xdkgxja4",
							"height": int64(2),
						},
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
				Results: map[string]any{
					"commits": []map[string]any{
						{
							// Composite head
							"cid":    "bafyreig3qosmew7pkq27dijjvwe35jjpvh3ed3f5dxpzemtqhw7xka7hga",
							"height": int64(3),
						},
						{
							// Composite head -1
							"cid":    "bafyreih5h6i6ohfsgrcjtg76iarebqcurpaft73gpobl2z2cfsvihsgdqu",
							"height": int64(2),
						},
						{
							// "Name" field head (unchanged from create)
							"cid":    "bafyreigtnj6ntulcilkmin4pgukjwv3nwglqpiiyddz3dyfexdbltze7sy",
							"height": int64(1),
						},
						{
							// "Age" field head
							"cid":    "bafyreiahq3xwdjmp2kq7jernt2axomiq3kuef2rik7k3fnn2pb242a5oha",
							"height": int64(3),
						},
						{
							// "Age" field head -1
							"cid":    "bafyreiale6qsjc7qewod3c6h2odwamfwcf7vt4zlqtw7ldcm57xdkgxja4",
							"height": int64(2),
						},
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
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": "bafyreiddiyec4bz2pqiav2bivqcqttr4kyniajrqxf66tybhq4cm36exi4",
						},
						{
							"cid": "bafyreicotst6miuynokequzsm7zjm42aw3zsfor7cvw7gja7hut3f5v6qq",
						},
						{
							"cid": "bafyreigiyb2tronlgaz4j5alh2a52gy7j5fi2ebvvf6r3dircvp6qkf4um",
						},
						{
							"cid": "bafyreif6dqbkr7t37jcjfxxrjnxt7cspxzvs7qwlbtjca57cc663he4s7e",
						},
						{
							"cid": "bafyreigtnj6ntulcilkmin4pgukjwv3nwglqpiiyddz3dyfexdbltze7sy",
						},
						{
							"cid": "bafyreia2vlbfkcbyogdjzmbqcjneabwwwtw7ti2xbd7yor5mbu2sk4pcoy",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
