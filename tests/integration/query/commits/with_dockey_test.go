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
						"Name":	"John",
						"Age":	21
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

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithDockey(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey",
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
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
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
						"cid": "bafybeihalsnuslda2ccygeq45nmrhspcg2yae56vsw7podra37a7ugemly",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithDockeyAndLinks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, with links",
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
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid":   "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeihalsnuslda2ccygeq45nmrhspcg2yae56vsw7podra37a7ugemly",
						"links": []map[string]any{
							{
								"cid":  "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
								"name": "Age",
							},
							{
								"cid":  "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
								"name": "Name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}

func TestQueryCommitsWithDockeyAndUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with dockey, multiple results",
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
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
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
						"cid":    "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
						"height": int64(1),
					},
					{
						"cid":    "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"height": int64(1),
					},
					{
						"cid":    "bafybeiat7moet2s4h3cfbektk2gtojzzbiinmuidlauyj6felzcmoud7fq",
						"height": int64(2),
					},
					{
						"cid":    "bafybeihalsnuslda2ccygeq45nmrhspcg2yae56vsw7podra37a7ugemly",
						"height": int64(1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
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
						commits(dockey: "bae-52b9170d-b77a-5887-b877-cbdbb99b009f") {
							cid
							links {
								cid
								name
							}
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeiepww5b67jrrliuiy27erfjuivwnjca5ptdpbxrrjrqkh3b2hckyy",
						"links": []map[string]any{
							{
								"cid":  "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
								"name": "_head",
							},
						},
					},
					{
						"cid":   "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
						"links": []map[string]any{},
					},
					{
						"cid":   "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
						"links": []map[string]any{},
					},
					{
						"cid": "bafybeiat7moet2s4h3cfbektk2gtojzzbiinmuidlauyj6felzcmoud7fq",
						"links": []map[string]any{
							{
								"cid":  "bafybeiepww5b67jrrliuiy27erfjuivwnjca5ptdpbxrrjrqkh3b2hckyy",
								"name": "Age",
							},
							{
								"cid":  "bafybeihalsnuslda2ccygeq45nmrhspcg2yae56vsw7podra37a7ugemly",
								"name": "_head",
							},
						},
					},
					{
						"cid": "bafybeihalsnuslda2ccygeq45nmrhspcg2yae56vsw7podra37a7ugemly",
						"links": []map[string]any{
							{
								"cid":  "bafybeihxvx3f7eejvco6zbxsidoeuph6ywpbo33lrqm3picna2aj7pdeiu",
								"name": "Age",
							},
							{
								"cid":  "bafybeih25dvtgei2bryhlz24tbyfdcni5di7akgcx24pezxts27wz7v454",
								"name": "Name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, []string{"users"}, test)
}
