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

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
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
						"cid":   "bafybeid4q6fhbbchwife54qqumb2rof6lui7d5njbkylkradmewqdibhjm",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeid435xjpnucmhshryyg3bfzf7be7hotq4m2kfw77yn7utd5yyimiq",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeic267ibnl45al5ekxpqorsbwv2xghsuxm4dpdi47ojhl7yuvdonuy",
						"links": []map[string]any{
							{
								"cid":  "bafybeid4q6fhbbchwife54qqumb2rof6lui7d5njbkylkradmewqdibhjm",
								"name": "age",
							},
							{
								"cid":  "bafybeid435xjpnucmhshryyg3bfzf7be7hotq4m2kfw77yn7utd5yyimiq",
								"name": "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
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
						"cid":    "bafybeidgiwk6kqpswcdnp5jmjgch6g2aqkrwqoiqcanxuxt3ne3huma7oi",
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
						"cid":    "bafybeidb4l2xwjdmcotorpivw3usowdx6rvinda2x26zakar2vm3r5tlse",
						"height": int64(2),
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
						"cid": "bafybeidgiwk6kqpswcdnp5jmjgch6g2aqkrwqoiqcanxuxt3ne3huma7oi",
						"links": []map[string]any{
							{
								"cid":  "bafybeid4q6fhbbchwife54qqumb2rof6lui7d5njbkylkradmewqdibhjm",
								"name": "_head",
							},
						},
					},
					{
						"cid":   "bafybeid4q6fhbbchwife54qqumb2rof6lui7d5njbkylkradmewqdibhjm",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeid435xjpnucmhshryyg3bfzf7be7hotq4m2kfw77yn7utd5yyimiq",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeidb4l2xwjdmcotorpivw3usowdx6rvinda2x26zakar2vm3r5tlse",
						"links": []map[string]any{
							{
								"cid":  "bafybeic267ibnl45al5ekxpqorsbwv2xghsuxm4dpdi47ojhl7yuvdonuy",
								"name": "_head",
							},
							{
								"cid":  "bafybeidgiwk6kqpswcdnp5jmjgch6g2aqkrwqoiqcanxuxt3ne3huma7oi",
								"name": "age",
							},
						},
					},
					{
						"cid": "bafybeic267ibnl45al5ekxpqorsbwv2xghsuxm4dpdi47ojhl7yuvdonuy",
						"links": []map[string]any{
							{
								"cid":  "bafybeid4q6fhbbchwife54qqumb2rof6lui7d5njbkylkradmewqdibhjm",
								"name": "age",
							},
							{
								"cid":  "bafybeid435xjpnucmhshryyg3bfzf7be7hotq4m2kfw77yn7utd5yyimiq",
								"name": "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
