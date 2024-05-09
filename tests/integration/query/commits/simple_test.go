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
						"cid": "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
					},
					{
						"cid": "bafyreidcls23tu7qwp4siw3avyb42eukovpxg6dqifqruvy5wyc6b2ovvq",
					},
					{
						"cid": "bafyreid7n6a673spwjwl3ogtuqmrba4i4ntjqvsu4l3spqe6qutdtnqwlq",
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
						"cid": "bafyreidoznu3dbvgngnfk3xjjgpmrpjs3wrpbc7ntl3rbacspeee5kp53a",
					},
					{
						"cid": "bafyreia5q5oya6vnv2kvffzfnl23762zvtqxmatjd5s3ldpwxfdo3aey6i",
					},
					{
						"cid": "bafyreibr7mv7b4kg4zym7fow6ljhl6kdzfgcvfi5q26ogz5gsev75ewxcq",
					},
					{
						"cid": "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
					},
					{
						"cid": "bafyreidcls23tu7qwp4siw3avyb42eukovpxg6dqifqruvy5wyc6b2ovvq",
					},
					{
						"cid": "bafyreid7n6a673spwjwl3ogtuqmrba4i4ntjqvsu4l3spqe6qutdtnqwlq",
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
						"cid":             "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
						"schemaVersionId": "bafkreicprhqxzlw3akyssz2v6pifwfueavp7jq2yj3dghapi3qcq6achs4",
					},
					{
						"cid":             "bafyreidcls23tu7qwp4siw3avyb42eukovpxg6dqifqruvy5wyc6b2ovvq",
						"schemaVersionId": "bafkreicprhqxzlw3akyssz2v6pifwfueavp7jq2yj3dghapi3qcq6achs4",
					},
					{
						"cid":             "bafyreid7n6a673spwjwl3ogtuqmrba4i4ntjqvsu4l3spqe6qutdtnqwlq",
						"schemaVersionId": "bafkreicprhqxzlw3akyssz2v6pifwfueavp7jq2yj3dghapi3qcq6achs4",
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
						"cid":          "bafyreictnkwvit6jp4mwhai3xp75nvtacxpq6zgbbjm55ylae3t6qshrze",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue(22),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(2),
						"links": []map[string]any{
							{
								"cid":  "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
								"name": "_head",
							},
						},
					},
					{
						"cid":          "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue(21),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreidcls23tu7qwp4siw3avyb42eukovpxg6dqifqruvy5wyc6b2ovvq",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue("John"),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "2",
						"fieldName":    "name",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreif3pvxatyqbmwllb7mcxvs734fgfmdkavbu6ambhay37w6vxjzkx4",
						"collectionID": int64(1),
						"delta":        []uint8(nil),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(2),
						"links": []map[string]any{
							{
								"cid":  "bafyreid7n6a673spwjwl3ogtuqmrba4i4ntjqvsu4l3spqe6qutdtnqwlq",
								"name": "_head",
							},
							{
								"cid":  "bafyreictnkwvit6jp4mwhai3xp75nvtacxpq6zgbbjm55ylae3t6qshrze",
								"name": "age",
							},
						},
					},
					{
						"cid":          "bafyreid7n6a673spwjwl3ogtuqmrba4i4ntjqvsu4l3spqe6qutdtnqwlq",
						"collectionID": int64(1),
						"delta":        []uint8(nil),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(1),
						"links": []map[string]any{
							{
								"cid":  "bafyreidcls23tu7qwp4siw3avyb42eukovpxg6dqifqruvy5wyc6b2ovvq",
								"name": "name",
							},
							{
								"cid":  "bafyreih5awhipv4pk7truqm3pyyhle7xersbiyzyyacud6c3f7urzutpui",
								"name": "age",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
