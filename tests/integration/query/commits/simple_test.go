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
						"cid": "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
					},
					{
						"cid": "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
					},
					{
						"cid": "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
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
						"cid": "bafyreihpasbgxcoxmzv5bp6euq3lbaoh5y5wjbbgfthtxqs3nppk36kebq",
					},
					{
						"cid": "bafyreihe3jydldbt7mvkiae6asrchdxajzkxwid6syi436nmrpcqhwt7xa",
					},
					{
						"cid": "bafyreihb5eo3luqoojztdmxtg3tdpvm6pc64mkyrzlefbdauker5qlnop4",
					},
					{
						"cid": "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
					},
					{
						"cid": "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
					},
					{
						"cid": "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithSchemaVersionIDField(t *testing.T) {
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
						"cid":             "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
						"schemaVersionId": "bafkreicprhqxzlw3akyssz2v6pifwfueavp7jq2yj3dghapi3qcq6achs4",
					},
					{
						"cid":             "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
						"schemaVersionId": "bafkreicprhqxzlw3akyssz2v6pifwfueavp7jq2yj3dghapi3qcq6achs4",
					},
					{
						"cid":             "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
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
						"cid":          "bafyreiay56ley5dvsptso37fsonfcrtbuphwlfhi67d2y52vzzexba6vua",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue(22),
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(2),
						"links": []map[string]any{
							{
								"cid":  "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
								"name": "_head",
							},
						},
					},
					{
						"cid":          "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue(21),
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue("John"),
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "2",
						"fieldName":    "name",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreicsavx5oblk6asfoqyssz4ge2gf5ekfouvi7o6l7adly275op5oje",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(2),
						"links": []map[string]any{
							{
								"cid":  "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
								"name": "_head",
							},
							{
								"cid":  "bafyreiay56ley5dvsptso37fsonfcrtbuphwlfhi67d2y52vzzexba6vua",
								"name": "age",
							},
						},
					},
					{
						"cid":          "bafyreihv7jqe32wsuff5vwzlp7izoo6pqg6kgqf5edknp3mqm3344gu35q",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(1),
						"links": []map[string]any{
							{
								"cid":  "bafyreic2sba5sffkfnt32wfeoaw4qsqozjb5acwwtouxuzllb3aymjwute",
								"name": "name",
							},
							{
								"cid":  "bafyreifzyy7bmpx2eywj4lznxzrzrvh6vrz6l7bhthkpexdq3wtho3vz6i",
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
