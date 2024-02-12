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

func TestQueryCommits(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query",
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
						commits {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeieikx6l2xead2dzsa5wwy5irxced2eddyq23jkp4csf5igoob7diq",
					},
					{
						"cid": "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
					},
					{
						"cid": "bafybeiggrv6gyhld2dbkspaxsenjejfhnk52pm4mlpyz2q6x4dlnaff2mu",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsMultipleDocs(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query, multiple docs",
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
						"name":	"Shahzad",
						"age":	28
					}`,
			},
			testUtils.Request{
				Request: `query {
						commits {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafybeifxmoq5yeukxrizqgncteztxao53d26ch6gqlyciakbrvlxwyob6a",
					},
					{
						"cid": "bafybeigkpb7jwgrsq3dq2nt2cnsarp3nq2ilefsuyz7jwcxfecot3qxwia",
					},
					{
						"cid": "bafybeidacwdvdzz7y5r6lqsjgs3kom2zjnlarw6r3rgxlj3iw4kck4mer4",
					},
					{
						"cid": "bafybeieikx6l2xead2dzsa5wwy5irxced2eddyq23jkp4csf5igoob7diq",
					},
					{
						"cid": "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
					},
					{
						"cid": "bafybeiggrv6gyhld2dbkspaxsenjejfhnk52pm4mlpyz2q6x4dlnaff2mu",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithSchemaVersionIdField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding schemaVersionId",
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
						commits {
							cid
							schemaVersionId
						}
					}`,
				Results: []map[string]any{
					{
						"cid":             "bafybeieikx6l2xead2dzsa5wwy5irxced2eddyq23jkp4csf5igoob7diq",
						"schemaVersionId": "bafkreidjvyxputjthx4wzyxtk33fce3shqguif3yhifykilybpn6canony",
					},
					{
						"cid":             "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
						"schemaVersionId": "bafkreidjvyxputjthx4wzyxtk33fce3shqguif3yhifykilybpn6canony",
					},
					{
						"cid":             "bafybeiggrv6gyhld2dbkspaxsenjejfhnk52pm4mlpyz2q6x4dlnaff2mu",
						"schemaVersionId": "bafkreidjvyxputjthx4wzyxtk33fce3shqguif3yhifykilybpn6canony",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithFieldNameField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding fieldName",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							fieldName
						}
					}
				`,
				Results: []map[string]any{
					{
						"fieldName": "age",
					},
					{
						"fieldName": "name",
					},
					{
						"fieldName": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithFieldNameFieldAndUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding fieldName",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							fieldName
						}
					}
				`,
				Results: []map[string]any{
					{
						"fieldName": "age",
					},
					{
						"fieldName": "age",
					},
					{
						"fieldName": "name",
					},
					{
						"fieldName": nil,
					},
					{
						"fieldName": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithFieldIDField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding fieldId",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							fieldId
						}
					}
				`,
				Results: []map[string]any{
					{
						"fieldId": "1",
					},
					{
						"fieldId": "2",
					},
					{
						"fieldId": "C",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithFieldIDFieldWithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple commits query yielding fieldId",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							fieldId
						}
					}
				`,
				Results: []map[string]any{
					{
						"fieldId": "1",
					},
					{
						"fieldId": "1",
					},
					{
						"fieldId": "2",
					},
					{
						"fieldId": "C",
					},
					{
						"fieldId": "C",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuery_CommitsWithAllFieldsWithUpdate_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							cid
							collectionID
							delta
							docID
							fieldId
							fieldName
							height
							links {
								cid
								name
							}
						}
					}
				`,
				Results: []map[string]any{
					{
						"cid":          "bafybeiddpjl27ulw2yo4ohup6gr2wob3pwagqw2rbeaxxodv4ljelnu7ve",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue(22),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(2),
						"links": []map[string]any{
							{
								"cid":  "bafybeieikx6l2xead2dzsa5wwy5irxced2eddyq23jkp4csf5igoob7diq",
								"name": "_head",
							},
						},
					},
					{
						"cid":          "bafybeieikx6l2xead2dzsa5wwy5irxced2eddyq23jkp4csf5igoob7diq",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue(21),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue("John"),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "2",
						"fieldName":    "name",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafybeiekajrgheumrgamrc4mprmm66ulp2qr75sviowfcriuviggokydbm",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(2),
						"links": []map[string]any{
							{
								"cid":  "bafybeiggrv6gyhld2dbkspaxsenjejfhnk52pm4mlpyz2q6x4dlnaff2mu",
								"name": "_head",
							},
							{
								"cid":  "bafybeiddpjl27ulw2yo4ohup6gr2wob3pwagqw2rbeaxxodv4ljelnu7ve",
								"name": "age",
							},
						},
					},
					{
						"cid":          "bafybeiggrv6gyhld2dbkspaxsenjejfhnk52pm4mlpyz2q6x4dlnaff2mu",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(1),
						"links": []map[string]any{
							{
								"cid":  "bafybeieikx6l2xead2dzsa5wwy5irxced2eddyq23jkp4csf5igoob7diq",
								"name": "age",
							},
							{
								"cid":  "bafybeiehcr3diremeja2ndk2osux647v5fc7s353h7pbvrnsagw4paugku",
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
