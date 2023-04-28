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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeidgiwk6kqpswcdnp5jmjgch6g2aqkrwqoiqcanxuxt3ne3huma7oi",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidb4l2xwjdmcotorpivw3usowdx6rvinda2x26zakar2vm3r5tlse",
						"height": int64(2),
					},
					{
						"cid":    "bafybeid4q6fhbbchwife54qqumb2rof6lui7d5njbkylkradmewqdibhjm",
						"height": int64(1),
					},
					{
						"cid":    "bafybeid435xjpnucmhshryyg3bfzf7be7hotq4m2kfw77yn7utd5yyimiq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeic267ibnl45al5ekxpqorsbwv2xghsuxm4dpdi47ojhl7yuvdonuy",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestQueryCommitsWithDockeyAndOrderHeightAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order height asc",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeid4q6fhbbchwife54qqumb2rof6lui7d5njbkylkradmewqdibhjm",
						"height": int64(1),
					},
					{
						"cid":    "bafybeid435xjpnucmhshryyg3bfzf7be7hotq4m2kfw77yn7utd5yyimiq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeic267ibnl45al5ekxpqorsbwv2xghsuxm4dpdi47ojhl7yuvdonuy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidgiwk6kqpswcdnp5jmjgch6g2aqkrwqoiqcanxuxt3ne3huma7oi",
						"height": int64(2),
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

func TestQueryCommitsWithDockeyAndOrderCidDesc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order cid desc",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: DESC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeidgiwk6kqpswcdnp5jmjgch6g2aqkrwqoiqcanxuxt3ne3huma7oi",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidb4l2xwjdmcotorpivw3usowdx6rvinda2x26zakar2vm3r5tlse",
						"height": int64(2),
					},
					{
						"cid":    "bafybeid4q6fhbbchwife54qqumb2rof6lui7d5njbkylkradmewqdibhjm",
						"height": int64(1),
					},
					{
						"cid":    "bafybeid435xjpnucmhshryyg3bfzf7be7hotq4m2kfw77yn7utd5yyimiq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeic267ibnl45al5ekxpqorsbwv2xghsuxm4dpdi47ojhl7yuvdonuy",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestQueryCommitsWithDockeyAndOrderCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, order cid asc",
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
						commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {cid: ASC}) {
							cid
							height
						}
					}`,
				Results: []map[string]any{
					{
						"cid":    "bafybeic267ibnl45al5ekxpqorsbwv2xghsuxm4dpdi47ojhl7yuvdonuy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeid435xjpnucmhshryyg3bfzf7be7hotq4m2kfw77yn7utd5yyimiq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeid4q6fhbbchwife54qqumb2rof6lui7d5njbkylkradmewqdibhjm",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidb4l2xwjdmcotorpivw3usowdx6rvinda2x26zakar2vm3r5tlse",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidgiwk6kqpswcdnp5jmjgch6g2aqkrwqoiqcanxuxt3ne3huma7oi",
						"height": int64(2),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestQueryCommitsWithDockeyAndOrderAndMultiUpdatesCidAsc(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple updates with order cid asc",
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
						 commits(dockey: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", order: {height: ASC}) {
							 cid
							 height
						 }
					 }`,
				Results: []map[string]any{
					{
						"cid":    "bafybeid4q6fhbbchwife54qqumb2rof6lui7d5njbkylkradmewqdibhjm",
						"height": int64(1),
					},
					{
						"cid":    "bafybeid435xjpnucmhshryyg3bfzf7be7hotq4m2kfw77yn7utd5yyimiq",
						"height": int64(1),
					},
					{
						"cid":    "bafybeic267ibnl45al5ekxpqorsbwv2xghsuxm4dpdi47ojhl7yuvdonuy",
						"height": int64(1),
					},
					{
						"cid":    "bafybeidgiwk6kqpswcdnp5jmjgch6g2aqkrwqoiqcanxuxt3ne3huma7oi",
						"height": int64(2),
					},
					{
						"cid":    "bafybeidb4l2xwjdmcotorpivw3usowdx6rvinda2x26zakar2vm3r5tlse",
						"height": int64(2),
					},
					{
						"cid":    "bafybeic77g2nj353n6djc6dcxexsi3uwfcaoappxlmvyhrt6lpiet6viry",
						"height": int64(3),
					},
					{
						"cid":    "bafybeif56zr6xacflpksrhuk2fe6kpe5s77d2txeb2j5begr6vlnuankye",
						"height": int64(3),
					},
					{
						"cid":    "bafybeid5rbs6rmsdvckroyycwmelixez7lvz6rbu76snoctolnj5rksf7m",
						"height": int64(4),
					},
					{
						"cid":    "bafybeidvvf53tnr4skn7ll4qygxaqft7ixkuunvlptvtf5tjxox6p7nwja",
						"height": int64(4),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
