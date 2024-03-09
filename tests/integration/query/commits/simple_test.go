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
						"cid": "bafybeicvpgfinf2m2jufbbcy5mhv6jca6in5k4fzx5op7xvvcmbp7sceaa",
					},
					{
						"cid": "bafybeib2espk2hq366wjnmazg45uvoswqbvf4plx7fgzayagxdn737onci",
					},
					{
						"cid": "bafybeigvpf62j7j2wbpid5iavzxielbhbsbbirmgzqkw3wpptdvysuztwi",
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
						"cid": "bafybeibhain2764v7eltfiam6dgwivfj56mvbme34nbdsdbndrsjkc2cje",
					},
					{
						"cid": "bafybeickrd5xayjhedyypf3yus55bkhpwd5dqlkdhivrcceexkpsgnic24",
					},
					{
						"cid": "bafybeieqyyprwrkbgyn7x4jkzmlnupnzpdymvbulef37brkzn7blqbe6l4",
					},
					{
						"cid": "bafybeicvpgfinf2m2jufbbcy5mhv6jca6in5k4fzx5op7xvvcmbp7sceaa",
					},
					{
						"cid": "bafybeib2espk2hq366wjnmazg45uvoswqbvf4plx7fgzayagxdn737onci",
					},
					{
						"cid": "bafybeigvpf62j7j2wbpid5iavzxielbhbsbbirmgzqkw3wpptdvysuztwi",
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
						"cid":             "bafybeicvpgfinf2m2jufbbcy5mhv6jca6in5k4fzx5op7xvvcmbp7sceaa",
						"schemaVersionId": "bafkreidqkjb23ngp34eebeaxiogrlogkpfz62vjb3clnnyvhlbgdaywkg4",
					},
					{
						"cid":             "bafybeib2espk2hq366wjnmazg45uvoswqbvf4plx7fgzayagxdn737onci",
						"schemaVersionId": "bafkreidqkjb23ngp34eebeaxiogrlogkpfz62vjb3clnnyvhlbgdaywkg4",
					},
					{
						"cid":             "bafybeigvpf62j7j2wbpid5iavzxielbhbsbbirmgzqkw3wpptdvysuztwi",
						"schemaVersionId": "bafkreidqkjb23ngp34eebeaxiogrlogkpfz62vjb3clnnyvhlbgdaywkg4",
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
						"cid":          "bafybeicwg56ddi7smy3j2kkv5y4yghvdrj3twqqafzdwtinbkw5mlpxwz4",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue(22),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(2),
						"links": []map[string]any{
							{
								"cid":  "bafybeicvpgfinf2m2jufbbcy5mhv6jca6in5k4fzx5op7xvvcmbp7sceaa",
								"name": "_head",
							},
						},
					},
					{
						"cid":          "bafybeicvpgfinf2m2jufbbcy5mhv6jca6in5k4fzx5op7xvvcmbp7sceaa",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue(21),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafybeib2espk2hq366wjnmazg45uvoswqbvf4plx7fgzayagxdn737onci",
						"collectionID": int64(1),
						"delta":        testUtils.CBORValue("John"),
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "2",
						"fieldName":    "name",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafybeidxnkwhuzmkdw5wuippru3tp74vcmz5jvcziambpjadxeathdh26a",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(2),
						"links": []map[string]any{
							{
								"cid":  "bafybeigvpf62j7j2wbpid5iavzxielbhbsbbirmgzqkw3wpptdvysuztwi",
								"name": "_head",
							},
							{
								"cid":  "bafybeicwg56ddi7smy3j2kkv5y4yghvdrj3twqqafzdwtinbkw5mlpxwz4",
								"name": "age",
							},
						},
					},
					{
						"cid":          "bafybeigvpf62j7j2wbpid5iavzxielbhbsbbirmgzqkw3wpptdvysuztwi",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(1),
						"links": []map[string]any{
							{
								"cid":  "bafybeicvpgfinf2m2jufbbcy5mhv6jca6in5k4fzx5op7xvvcmbp7sceaa",
								"name": "age",
							},
							{
								"cid":  "bafybeib2espk2hq366wjnmazg45uvoswqbvf4plx7fgzayagxdn737onci",
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
