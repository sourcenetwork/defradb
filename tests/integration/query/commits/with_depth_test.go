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
						"cid": "bafybeid4q6fhbbchwife54qqumb2rof6lui7d5njbkylkradmewqdibhjm",
					},
					{
						"cid": "bafybeid435xjpnucmhshryyg3bfzf7be7hotq4m2kfw77yn7utd5yyimiq",
					},
					{
						"cid": "bafybeic267ibnl45al5ekxpqorsbwv2xghsuxm4dpdi47ojhl7yuvdonuy",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
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
						"cid":    "bafybeidgiwk6kqpswcdnp5jmjgch6g2aqkrwqoiqcanxuxt3ne3huma7oi",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeid435xjpnucmhshryyg3bfzf7be7hotq4m2kfw77yn7utd5yyimiq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidb4l2xwjdmcotorpivw3usowdx6rvinda2x26zakar2vm3r5tlse",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
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
						"cid":    "bafybeic77g2nj353n6djc6dcxexsi3uwfcaoappxlmvyhrt6lpiet6viry",
						"height": int64(3),
					},
					{
						// Composite head -1
						"cid":    "bafybeidgiwk6kqpswcdnp5jmjgch6g2aqkrwqoiqcanxuxt3ne3huma7oi",
						"height": int64(2),
					},
					{
						// "Name" field head (unchanged from create)
						"cid":    "bafybeid435xjpnucmhshryyg3bfzf7be7hotq4m2kfw77yn7utd5yyimiq",
						"height": int64(1),
					},
					{
						// "Age" field head
						"cid":    "bafybeif56zr6xacflpksrhuk2fe6kpe5s77d2txeb2j5begr6vlnuankye",
						"height": int64(3),
					},
					{
						// "Age" field head -1
						"cid":    "bafybeidb4l2xwjdmcotorpivw3usowdx6rvinda2x26zakar2vm3r5tlse",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
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
						"cid": "bafybeihginsps2jys6ba2liadfj3vvi7lhvtqzt73jm2xttizqo6wz6h3a",
					},
					{
						"cid": "bafybeihihnjplrkzrn6hg4zttfcbrhhxbl42qxoxaeapy4qcnunb7fmqmq",
					},
					{
						"cid": "bafybeig63b2mofkdljhclpmte55cza2mvcsjazccnxdydgsrekodrmb6e4",
					},
					{
						"cid": "bafybeid4q6fhbbchwife54qqumb2rof6lui7d5njbkylkradmewqdibhjm",
					},
					{
						"cid": "bafybeid435xjpnucmhshryyg3bfzf7be7hotq4m2kfw77yn7utd5yyimiq",
					},
					{
						"cid": "bafybeic267ibnl45al5ekxpqorsbwv2xghsuxm4dpdi47ojhl7yuvdonuy",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
