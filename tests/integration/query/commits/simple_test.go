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
						"cid": "bafybeiev2thtqxttuhr3aq5dvyb3aif4cey7werksskw5xuetmwxjxi7ty",
					},
					{
						"cid": "bafybeifos5iir63tmp3bdoj7zr5aand4ud2tf2qfnjlh6nvrzw3knkewuy",
					},
					{
						"cid": "bafybeicjb4x47xk6koh4uhgokhjr5zbg3bhbcfoa4um4vdktnbhrsx6d2a",
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
						"cid":             "bafybeieoeoset5itv7alud2yzjmq6dqizymdwdmlvyxam2uxe4lfexooaq",
						"schemaVersionId": "bafkreibpk53kumv6q3kkc3hz2y57tnbgizpodk3vleyc3h5muv6vm4pdoe",
					},
					{
						"cid":             "bafybeih4plbb3rinhqvn663ssfwhnujdbnbjistymzowsry5nvmxchmqny",
						"schemaVersionId": "bafkreibpk53kumv6q3kkc3hz2y57tnbgizpodk3vleyc3h5muv6vm4pdoe",
					},
					{
						"cid":             "bafybeic2zvs2beirqmgd45myszkqwj32w3oyduolugkxv4gxxph4c4mzva",
						"schemaVersionId": "bafkreibpk53kumv6q3kkc3hz2y57tnbgizpodk3vleyc3h5muv6vm4pdoe",
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
						"cid":          "bafybeia7qkfbfm4jijlkqs6uxziie2v57nin5gaa3afnpkruw352mmrt4q",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue(22),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(2),
						"links": []map[string]any{
							{
								"cid":  "bafybeieoeoset5itv7alud2yzjmq6dqizymdwdmlvyxam2uxe4lfexooaq",
								"name": "_head",
							},
						},
					},
					{
						"cid":          "bafybeieoeoset5itv7alud2yzjmq6dqizymdwdmlvyxam2uxe4lfexooaq",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue(21),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafybeih4plbb3rinhqvn663ssfwhnujdbnbjistymzowsry5nvmxchmqny",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue("John"),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "2",
						"fieldName":    "name",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafybeid4fh7ggr2wgema6b5hrqroimcso3vxyous3oyck5c66vm72br7z4",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(2),
						"links": []map[string]any{
							{
								"cid":  "bafybeic2zvs2beirqmgd45myszkqwj32w3oyduolugkxv4gxxph4c4mzva",
								"name": "_head",
							},
							{
								"cid":  "bafybeia7qkfbfm4jijlkqs6uxziie2v57nin5gaa3afnpkruw352mmrt4q",
								"name": "age",
							},
						},
					},
					{
						"cid":          "bafybeic2zvs2beirqmgd45myszkqwj32w3oyduolugkxv4gxxph4c4mzva",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(1),
						"links": []map[string]any{
							{
								"cid":  "bafybeieoeoset5itv7alud2yzjmq6dqizymdwdmlvyxam2uxe4lfexooaq",
								"name": "age",
							},
							{
								"cid":  "bafybeih4plbb3rinhqvn663ssfwhnujdbnbjistymzowsry5nvmxchmqny",
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
